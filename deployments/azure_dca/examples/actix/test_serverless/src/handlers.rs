use actix_web::{HttpResponse, web::Data, Error};

use crate::{app_state::AppState};

pub(crate) async fn handle_get_secret_message(
    data: Data<AppState>,
) -> Result<HttpResponse, Error>  {
    let message = data.into_inner().message.clone();
    Ok(HttpResponse::Ok().body(message))
}
