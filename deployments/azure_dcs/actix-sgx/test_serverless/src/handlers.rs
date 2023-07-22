use actix_web::{HttpResponse,Error};

pub(crate) async fn handle_hello_world(
) -> Result<HttpResponse, Error>  {
    Ok(HttpResponse::Ok().body("Hello world!"))
}
