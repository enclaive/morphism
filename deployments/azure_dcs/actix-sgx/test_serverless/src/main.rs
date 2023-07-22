use actix_web::{HttpServer, web, App};
use openssl::ssl::{SslAcceptor, SslMethod, SslFiletype};
use handlers::handle_hello_world;
mod handlers;

#[actix_web::main] 
async fn main() -> std::io::Result<()> {
    println!("Web server started...");
    let mut builder = SslAcceptor::mozilla_intermediate(SslMethod::tls()).unwrap();
    builder.set_private_key_file("/secrets/tmp/key.pem", SslFiletype::PEM).unwrap();
    builder.set_certificate_chain_file("/secrets/tmp/cert.pem").unwrap();
    HttpServer::new(move || {
        App::new()
        .service(web::resource("/hello_world").route(web::get().to(handle_hello_world)))
    })
    .bind_openssl("0.0.0.0:8080", builder)?
    .run()
    .await
}