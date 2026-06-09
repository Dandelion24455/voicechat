import { config } from './config.js';

export class RoomManager {
  constructor(getWorker) {
    this.getWorker = getWorker;
    this.rooms = new Map();
  }

  getOrCreate(roomId) {
    if (!this.rooms.has(roomId)) {
      this.rooms.set(roomId, new Room(roomId, this.getWorker()));
    }
    return this.rooms.get(roomId);
  }

  closeRoom(roomId) {
    const room = this.rooms.get(roomId);
    if (room) {
      room.close();
      this.rooms.delete(roomId);
    }
  }
}

class Room {
  constructor(id, worker) {
    this.id = id;
    this.worker = worker;
    this.router = null;
    this.peers = new Map();
    this._init = null;
    console.log(`[room ${id}] created`);
  }

  async init() {
    if (this._init) return this._init;
    this._init = (async () => {
      this.router = await this.worker.createRouter({
        mediaCodecs: config.mediasoup.router.mediaCodecs,
      });
      console.log(`[room ${this.id}] router ready`);
    })();
    return this._init;
  }

  addPeer(userId, ws, username) {
    const peer = { userId, ws, username, producerTransport: null, consumerTransport: null, producers: new Map(), consumers: new Map() };
    this.peers.set(userId, peer);
    console.log(`[room ${this.id}] peer joined: ${username} (${this.peers.size} total)`);
    return peer;
  }

  removePeer(userId) {
    const peer = this.peers.get(userId);
    if (!peer) return;
    for (const [pid, producer] of peer.producers) {
      producer.close();
      this.broadcast({ type: 'producerClosed', producerId: pid }, userId);
    }
    peer.producerTransport?.close();
    peer.consumerTransport?.close();
    for (const [, consumer] of peer.consumers) {
      consumer.close();
    }
    this.peers.delete(userId);
    console.log(`[room ${this.id}] peer left: ${peer.username} (${this.peers.size} total)`);
    if (this.peers.size === 0) {
      this.close();
    }
  }

  close() {
    this.router?.close();
    this.peers.clear();
    console.log(`[room ${this.id}] closed`);
  }

  send(userId, msg) {
    const peer = this.peers.get(userId);
    if (peer && peer.ws.readyState === 1) {
      peer.ws.send(JSON.stringify(msg));
    }
  }

  broadcast(msg, excludeUserId) {
    for (const [uid, peer] of this.peers) {
      if (uid !== excludeUserId && peer.ws.readyState === 1) {
        peer.ws.send(JSON.stringify(msg));
      }
    }
  }

  async handleMessage(userId, msg) {
    const peer = this.peers.get(userId);
    if (!peer) return;

    try {
      switch (msg.type) {
        case 'createProducerTransport':
          await this.onCreateProducerTransport(peer);
          break;
        case 'connectProducerTransport':
          await this.onConnectProducerTransport(peer, msg);
          break;
        case 'produce':
          await this.onProduce(peer, msg, userId);
          break;
        case 'createConsumerTransport':
          await this.onCreateConsumerTransport(peer);
          break;
        case 'connectConsumerTransport':
          await this.onConnectConsumerTransport(peer, msg);
          break;
        case 'consume':
          await this.onConsume(peer, msg);
          break;
        case 'resume':
          await this.onResume(peer, msg);
          break;
        default:
          this.send(userId, { type: 'error', message: 'unknown message type: ' + msg.type });
      }
    } catch (err) {
      console.error(`[room ${this.id}] error handling ${msg.type} from ${userId}:`, err.message);
      this.send(userId, { type: 'error', message: err.message });
    }
  }

  async onCreateProducerTransport(peer) {
    const transport = await this.router.createWebRtcTransport(config.mediasoup.webRtcTransport);
    peer.producerTransport = transport;

    transport.observer.on('close', () => { peer.producerTransport = null; });

    this.send(peer.userId, {
      type: 'producerTransportCreated',
      id: transport.id,
      iceParameters: transport.iceParameters,
      iceCandidates: transport.iceCandidates,
      dtlsParameters: transport.dtlsParameters,
    });
  }

  async onConnectProducerTransport(peer, msg) {
    const transport = peer.producerTransport;
    if (!transport) throw new Error('no producer transport');
    await transport.connect({ dtlsParameters: msg.dtlsParameters });
    this.send(peer.userId, { type: 'connectProducerTransportAck' });
  }

  async onProduce(peer, msg, userId) {
    const transport = peer.producerTransport;
    if (!transport) throw new Error('no producer transport');

    const producer = await transport.produce({ kind: msg.kind, rtpParameters: msg.rtpParameters });
    peer.producers.set(producer.id, producer);

    producer.observer.on('close', () => { peer.producers.delete(producer.id); });

    this.send(peer.userId, { type: 'produced', id: producer.id });

    this.broadcast({
      type: 'newProducer',
      producerId: producer.id,
      userId,
      username: peer.username,
    }, userId);
  }

  async onCreateConsumerTransport(peer) {
    const transport = await this.router.createWebRtcTransport(config.mediasoup.webRtcTransport);
    peer.consumerTransport = transport;

    transport.observer.on('close', () => { peer.consumerTransport = null; });

    this.send(peer.userId, {
      type: 'consumerTransportCreated',
      id: transport.id,
      iceParameters: transport.iceParameters,
      iceCandidates: transport.iceCandidates,
      dtlsParameters: transport.dtlsParameters,
    });
  }

  async onConnectConsumerTransport(peer, msg) {
    const transport = peer.consumerTransport;
    if (!transport) throw new Error('no consumer transport');
    await transport.connect({ dtlsParameters: msg.dtlsParameters });
    this.send(peer.userId, { type: 'connectConsumerTransportAck' });
  }

  async onConsume(peer, msg) {
    const transport = peer.consumerTransport;
    if (!transport) throw new Error('no consumer transport');

    let producer = null;
    for (const [, p] of this.peers) {
      if (p.producers.has(msg.producerId)) {
        producer = p.producers.get(msg.producerId);
        break;
      }
    }
    if (!producer) {
      this.send(peer.userId, { type: 'producerClosed', producerId: msg.producerId });
      return;
    }

    const consumer = await transport.consume({
      producerId: msg.producerId,
      rtpCapabilities: msg.rtpCapabilities,
      paused: false,
    });

    peer.consumers.set(consumer.id, consumer);
    consumer.observer.on('close', () => { peer.consumers.delete(consumer.id); });

    this.send(peer.userId, {
      type: 'consumed',
      id: consumer.id,
      producerId: msg.producerId,
      kind: consumer.kind,
      rtpParameters: consumer.rtpParameters,
    });
  }

  async onResume(peer, msg) {
    const consumer = peer.consumers.get(msg.consumerId);
    if (consumer) {
      await consumer.resume();
      this.send(peer.userId, { type: 'resumed', consumerId: msg.consumerId });
    }
  }
}
