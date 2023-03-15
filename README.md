# Audit/Logger Webhook

`webhook` listens for incoming events from MinIO server and logs these events to a log file.

Usage:
```
webhook --log-file <logfile>
```

Environment only settings:

| ENV                | Description                                                        |
|--------------------|--------------------------------------------------------------------|
| WEBHOOK_AUTH_TOKEN | Authorization token optional to authenticate/trust incoming events |

The webhook service can be setup as a systemd service using the `webhook.service` shipped with
this project

To send Audit logs from MinIO server, please configure MinIO using the command:
```
mc admin config set myminio audit_webhook endpoint=http://webhookendpoint:8080 auth_token=webhooksecret
```

To send server logs from MinIO server, please configure MinIO using the command:
```
mc admin config set myminio logger_webhook endpoint=http://webhookendpoint:8081 auth_token=webhooksecret
```

> NOTE: audit_webhook and logger_webhook should *not* be configured to send events to the same webhook instance.

Logs can be rotated using the standard logrotate tool. You can provide the postrotate command such that
webhook writes to a new log file after log rotation.
```
postrotate
	systemctl reload webhook
endscript
```
