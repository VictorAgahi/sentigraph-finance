use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct AppConfig {
    pub redis_url: String,
    pub binance_ws_url: String,
    #[serde(default = "default_log_level")]
    pub rust_log: String,
}

fn default_log_level() -> String {
    "info".to_string()
}

impl AppConfig {
    pub fn load() -> Self {
        dotenvy::dotenv().ok();

        envy::from_env::<AppConfig>()
            .expect("FATAL: Missing required environment variables.")
    }
}
