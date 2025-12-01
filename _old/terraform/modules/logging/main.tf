resource "kubernetes_config_map" "fluentd" {
  count = var.fluentd_enabled ? 1 : 0

  metadata {
    name      = "${var.release_name}-fluentd-config"
    namespace = var.namespace
  }

  data = {
    "fluent.conf" = <<-EOT
      <source>
        @type tail
        path /var/log/containers/*.log
        pos_file /var/log/fluentd-containers.log.pos
        tag kubernetes.*
        read_from_head true
        <parse>
          @type json
          time_format %Y-%m-%dT%H:%M:%S.%NZ
        </parse>
      </source>

      <filter kubernetes.**>
        @type kubernetes_metadata
      </filter>

      <match **>
        @type elasticsearch
        host elasticsearch.${var.namespace}.svc.cluster.local
        port 9200
        logstash_format true
        logstash_prefix fluentd
        <buffer>
          @type file
          path /var/log/fluentd-buffers/kubernetes.system.buffer
          flush_mode interval
          retry_type exponential_backoff
          flush_interval 5s
          retry_max_interval 30
          chunk_limit_size 2M
          queue_limit_length 8
          overflow_action block
        </buffer>
      </match>
    EOT
  }
}

resource "kubernetes_daemonset" "fluentd" {
  count = var.fluentd_enabled ? 1 : 0

  metadata {
    name      = "${var.release_name}-fluentd"
    namespace = var.namespace
    labels = {
      app       = "fluentd"
      component = "logging"
    }
  }

  spec {
    selector {
      match_labels = {
        app = "fluentd"
      }
    }

    template {
      metadata {
        labels = {
          app       = "fluentd"
          component = "logging"
        }
      }

      spec {
        service_account_name = var.service_account_name

        container {
          name  = "fluentd"
          image = "fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch"

          env {
            name  = "FLUENT_ELASTICSEARCH_HOST"
            value = "elasticsearch.${var.namespace}.svc.cluster.local"
          }

          env {
            name  = "FLUENT_ELASTICSEARCH_PORT"
            value = "9200"
          }

          resources {
            limits = {
              memory = "200Mi"
              cpu    = "100m"
            }
            requests = {
              memory = "100Mi"
              cpu    = "50m"
            }
          }

          volume_mount {
            name       = "varlog"
            mount_path = "/var/log"
          }

          volume_mount {
            name       = "varlibdockercontainers"
            mount_path = "/var/lib/docker/containers"
            read_only  = true
          }

          volume_mount {
            name       = "config"
            mount_path = "/fluentd/etc/fluent.conf"
            sub_path   = "fluent.conf"
          }
        }

        volume {
          name = "varlog"
          host_path {
            path = "/var/log"
          }
        }

        volume {
          name = "varlibdockercontainers"
          host_path {
            path = "/var/lib/docker/containers"
          }
        }

        volume {
          name = "config"
          config_map {
            name = kubernetes_config_map.fluentd[0].metadata[0].name
          }
        }
      }
    }
  }
}
