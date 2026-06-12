import mediasoup from 'mediasoup';
import os from 'os';
import { WebSocketServer } from 'ws';
import jwt from 'jsonwebtoken';
import { config } from './config.js';
import { RoomManager } from './room.js';

async function main() {
  const numWorkers = Math.min(os.cpus().length, config.maxWorkers || 4);
  console.log(`creating ${numWorkers} mediasoup workers...`);

  const workers = [];
  for (let i = 0; i < numWorkers; i++) {
    const worker = await mediasoup.createWorker({
      rtcMinPort: config.mediasoup.worker.rtcMinPort,
      rtcMaxPort: config.mediasoup.worker.rtcMaxPort,
      logLevel: config.mediasoup.worker.logLevel,
    });
    worker.on('died', () => {
      console.error(`mediasoup worker ${i} died, exiting`);
      process.exit(1);
    });
    workers.push(worker);
    console.log(`worker ${i} ready`);
  }
  console.log(`all ${numWorkers} workers ready (UDP ${config.mediasoup.worker.rtcMinPort}-${config.mediasoup.worker.rtcMaxPort})`);

  let rrIndex = 0;
  function getWorker() {
    return workers[rrIndex++ % workers.length];
  }

  const roomManager = new RoomManager(getWorker);

  const wss = new WebSocketServer({
    port: config.port,
    verifyClient: (info, cb) => {
      const url = new URL(info.req.url, 'http://localhost');
      const token = url.searchParams.get('token');
      const roomId = url.searchParams.get('roomId');

      if (!token || !roomId) {
        cb(false, 401, 'missing token or roomId');
        return;
      }

      try {
        const claims = jwt.verify(token, config.jwtSecret);
        info.req._auth = {
          userId: claims.user_id,
          playerId: claims.player_id,
          username: claims.username,
          roomId,
        };
        cb(true);
      } catch (err) {
        console.error('jwt verify failed:', err.message);
        cb(false, 401, 'invalid token');
      }
    },
  });

  wss.on('connection', async (ws, req) => {
    const { userId, playerId, username, roomId } = req._auth;
    console.log(`[ws] ${username} (${playerId}) connecting to room ${roomId}`);

    const room = roomManager.getOrCreate(roomId);
    await room.init();
    room.addPeer(userId, ws, username);

    ws.send(JSON.stringify({ type: 'welcome', userId, username, roomId, playerId }));
    ws.send(JSON.stringify({
      type: 'routerRtpCapabilities',
      routerRtpCapabilities: room.router.rtpCapabilities,
    }));
    // Notify new peer about existing producers in the room.
    room.sendExistingProducers(userId);

    ws.on('message', (data) => {
      let msg;
      try {
        msg = JSON.parse(data.toString());
      } catch {
        return;
      }
      room.handleMessage(userId, msg);
    });

    ws.on('close', () => {
      room.removePeer(userId);
      if (room.peers.size === 0) {
        roomManager.closeRoom(roomId);
      }
    });

    ws.on('error', (err) => {
      console.error(`[ws] error from ${username}:`, err.message);
    });
  });

  console.log(`mediasoup signaling server on :${config.port}`);

  process.on('SIGTERM', () => { wss.close(); workers.forEach(w => w.close()); process.exit(0); });
  process.on('SIGINT', () => { wss.close(); workers.forEach(w => w.close()); process.exit(0); });
}

main().catch(err => {
  console.error('fatal:', err);
  process.exit(1);
});
