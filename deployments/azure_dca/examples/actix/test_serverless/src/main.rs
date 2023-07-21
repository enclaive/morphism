use actix_web::{HttpServer, web::{self, Data}, App};
use app_state::AppState;
use handlers::handle_get_secret_message;
mod app_state;
mod models;
mod handlers;

#[actix_web::main] 
async fn main() -> std::io::Result<()> {
    println!("Web server started...");
    let data = Data::new(AppState::generate().await);
    HttpServer::new(move || {
        App::new()
        .app_data(data.clone())
        .service(web::resource("/read_secret").route(web::get().to(handle_get_secret_message)))
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
