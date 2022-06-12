# device-gpio-alarm

This service creates alerts based on GPIO sensor input.  It uses https://github.com/warthog618/gpiod and listens in 'pull up' mode so that it can get alerts onto the edgex message bus as soon as possible.

## commands

The only command supported is `Alert`- same as the events pushed off.  Alerts will be for a configured duration and read commands will return true if one is active.

Write commands will be considered an acknowledgement and will sound the associated alarm if the alert is still active.  If `RequireAck` is configured to true the alarm will not be triggered until acknowledgement is received (via a PUT to "Alert")