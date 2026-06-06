use serde::Serialize;
use tauri::{
    menu::{MenuBuilder, MenuItemBuilder},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
    Emitter, Manager, Runtime, Window,
};

#[derive(Clone, Serialize)]
struct TrayEvent {
    action: String,
}

fn toggle_window<R: Runtime>(app: &tauri::AppHandle<R>) {
    if let Some(window) = app.get_webview_window("main") {
        if window.is_visible().unwrap_or(false) {
            let _ = window.hide();
        } else {
            let _ = window.show();
            let _ = window.set_focus();
        }
    }
}

#[tauri::command]
fn get_api_url() -> String {
    std::env::var("API_URL").unwrap_or_else(|_| "/api".into())
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let toggle = MenuItemBuilder::with_id("toggle", "Show/Hide").build(app)?;
            let mute = MenuItemBuilder::with_id("mute", "Mute").build(app)?;
            let quit = MenuItemBuilder::with_id("quit", "Quit").build(app)?;
            let menu = MenuBuilder::new(app).item(&toggle).item(&mute).item(&quit).build()?;

            let _tray = TrayIconBuilder::new()
                .menu(&menu)
                .tooltip("VoiceChat")
                .on_menu_event(move |app, event| match event.id().as_ref() {
                    "toggle" => toggle_window(app),
                    "mute" => {
                        let _ = app.emit("tray-action", TrayEvent { action: "mute".into() });
                    }
                    "quit" => app.exit(0),
                    _ => {}
                })
                .on_tray_icon_event(|tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        toggle_window(tray.app_handle());
                    }
                })
                .build(app)?;

            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                api.prevent_close();
                let _ = window.hide();
            }
        })
        .invoke_handler(tauri::generate_handler![get_api_url])
        .run(tauri::generate_context!())
        .expect("error while running VoiceChat");
}
