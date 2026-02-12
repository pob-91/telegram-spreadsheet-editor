# redis
resource "kubernetes_deployment" "finances_valkey" {
  metadata {
    name      = "finances-valkey"
    namespace = "default"
  }

  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "finances-valkey"
      }
    }

    template {
      metadata {
        labels = {
          app = "finances-valkey"
        }
      }

      spec {
        container {
          name  = "finances-valkey"
          image = "valkey:9-alpine"
          port {
            container_port = 6379
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "finances_valkey" {
  metadata {
    name      = "finances-valkey"
    namespace = "default"
  }

  spec {
    selector = {
      app = "finances-valkey"
    }

    port {
      port        = 6379
      target_port = 6379
    }

    type = "ClusterIP"
  }
}

# app

resource "kubernetes_config_map" "finances_editor_config" {
  metadata {
    name      = "finances-editor-config"
    namespace = "default"
  }

  data = {
    "config.yaml" = file("${path.module}/config.yaml")
  }
}

resource "kubernetes_deployment" "finances_editor" {
  metadata {
    name      = "finances-editor"
    namespace = "default"
  }

  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "finances-editor"
      }
    }

    template {
      metadata {
        labels = {
          app = "finances-editor"
        }
      }

      spec {
        container {
          name  = "finances-editor"
          image = "ghcr.io/pob-91/telegram-spreadsheet-editor:${var.finances_editor_version}"

          env {
            name  = "BASIC_AUTH_USER"
            value = "admin"
          }
          env {
            name  = "BASIC_AUTH_PASSWORD"
            value = var.finances_editor_basic_auth_password
          }
          env {
            name  = "SHEET_BASE_URL"
            value = "https://nextcloud_url/remote.php/dav/files/admin"
          }
          env {
            name  = "XLSX_FILE_PATH"
            value = "Documents/Finances.xlsx"
          }
          env {
            name  = "KEY_COLUMN"
            value = "D"
          }
          env {
            name  = "VALUE_COLUMN"
            value = "E"
          }
          env {
            name  = "START_ROW"
            value = "2"
          }
          env {
            name  = "TELEGRAM_BOT_TOKEN"
            value = var.finances_editor_telegram_token
          }
          dynamic "env" {
            for_each = length(var.finances_editor_allowed_telegram_users) > 0 ? [1] : []
            content {
              name  = "TELEGRAM_ALLOWED_USERS"
              value = join(",", var.finances_editor_allowed_telegram_users)
            }
          }
          env {
            name  = "SERVICE_HOST"
            value = "https://my-cool-domain.com"
          }
          env {
            name  = "VALKEY_HOST"
            value = "${kubernetes_service.finances_valkey.metadata[0].name}:6379"
          }
          env {
            name  = "LOG_LEVEL"
            value = "Warning"
          }

          port {
            container_port = 8080
          }

          volume_mount {
            name       = "finances-editor-config"
            mount_path = "/home/nonroot/config.yaml"
            sub_path   = "config.yaml"
            read_only  = true
          }
        }

        volume {
          name = "finances-editor-config"
          config_map {
            name = kubernetes_config_map.finances_editor_config.metadata[0].name
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "finances_editor" {
  metadata {
    name      = "finances-editor"
    namespace = "default"
  }

  spec {
    selector = {
      app = "finances-editor"
    }

    port {
      port        = 8080
      target_port = 8080
    }

    type = "ClusterIP"
  }
}

resource "kubernetes_manifest" "finances_editor_ingress" {
  manifest = {
    apiVersion = "networking.k8s.io/v1"
    kind       = "Ingress"
    metadata = {
      name      = "finances-editor"
      namespace = "default"
      annotations = {
        "kubernetes.io/ingress.class" = "traefik"
      }
    }
    spec = {
      rules = [
        {
          host = "my-cool-domain.com"
          http = {
            paths = [
              {
                path     = "/"
                pathType = "Prefix"
                backend = {
                  service = {
                    name = "finances-editor"
                    port = { number = 8080 }
                  }
                }
              }
            ]
          }
        }
      ]
    }
  }
}
