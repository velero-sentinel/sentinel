# Velero Sentinel

[![Stage: pre-alpha](https://img.shields.io/badge/Stage-pre--alpha-yellow)][wp:stage] [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=velero-sentinel_sentinel&metric=alert_status)](https://sonarcloud.io/dashboard?id=velero-sentinel_sentinel) [![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=velero-sentinel_sentinel&metric=security_rating)](https://sonarcloud.io/dashboard?id=velero-sentinel_sentinel) [![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=velero-sentinel_sentinel&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=velero-sentinel_sentinel) [![Coverage](https://sonarcloud.io/api/project_badges/measure?project=velero-sentinel_sentinel&metric=coverage)](https://sonarcloud.io/dashboard?id=velero-sentinel_sentinel)

> Sentinel is currently WIP and not ready to use for production &ndash; yet.

Velero Sentinel is a small service monitoring [Velero backups][velero:backups].
It will send notifications if a backup fails partially or completely.

## Event types

There are two event types: warnings and alerts. A partially failed backup will trigger a warning, while a failed backup will trigger an alert.

## Notification Channels

There are several notification channels planned.

- [x] Logs.
- [x] Webhooks with a template based content. So in theory, it should be possible to use this generic webhook for Slack, Rocketchat, Teams and whatnot.
- [ ] OpsGenie
- [ ] AMQP
- [ ] NATS

No SMTP? I have not planned for it as of now, since I think SMTP is an utterly useless protocol for reliable alerting, due to its nature. However, if demand should prove high enough and somebody is willing to get me something of my [Amazon wishlist][wishlist], I will implement it. Of course, as always, pull requests are welcome.

[![SonarCloud](https://sonarcloud.io/images/project_badges/sonarcloud-orange.svg)](https://sonarcloud.io/dashboard?id=velero-sentinel_sentinel)

[wp:stage]: https://en.wikipedia.org/wiki/Software_release_life_cycle#Pre-alpha
[velero:backups]: https://velero.io/docs/main/ "\"About Velero\" on velero.io"
[wishlist]: https://www.amazon.de/hz/wishlist/ls/1ELWVEKV9NLYP?ref_=wl_share "My wishlist on Amazon.DE"
