resource "immich_system_config" "example" {
  password_login = {
    enabled = true
  }

  oauth = {
    enabled       = true
    issuer_url    = "https://keycloak.example.com/realms/immich"
    client_id     = "immich-client"
    client_secret = "your-client-secret"
    scope         = "openid profile email"
    button_text   = "Login with Keycloak"
    auto_register = true
  }

  storage_template = {
    template = "{{y}}/{{y}}-{{m}}-{{d}}/{{filename}}"
  }

  machine_learning = {
    enabled    = true
    url        = "http://immich-machine-learning:3003"
    clip_model = "ViT-L-14__openai"
  }

  notifications = {
    smtp = {
      enabled  = true
      host     = "smtp.example.com"
      port     = 587
      username = "user@example.com"
      password = "secure-smtp-password"
      from     = "immich@example.com"
      secure   = false
    }
  }

  templates = {
    email = {
      welcome_template = "Welcome to Immich, {{name}}!"
    }
  }
}
