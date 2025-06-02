resource "cloudflare_queue" "%[1]s_queue" {
  account_id = "%[2]s"
  queue_name = "%[4]s"
}

resource "cloudflare_r2_bucket" "%[1]s_bucket" {
  account_id = "%[2]s"
  name       = "%[3]s"
}

resource "cloudflare_r2_bucket_event_notification" "%[1]s" {
  account_id   = "%[2]s"
  bucket_name  = cloudflare_r2_bucket.%[1]s_bucket.name
  queue_id     = cloudflare_queue.%[1]s_queue.queue_id
  jurisdiction = "default"

  rules = [{
    actions = ["PutObject"]
  }]
  depends_on = [ cloudflare_queue.%[1]s_queue ]
}
