# device-rpio-alarm

This service creates alerts based on GPIO sensor input.  It reads the memory associated with the sensor pin constantly so that it can get alerts onto the edgex message bus as soon as possible.

## commands

The only command supported is `Alert`- same as the events pushed off.  Alerts will be for a configured duration and read commands will return true if one is active.

Write commands will be considered an acknowledgement and will sound the associated alarm if the alert is still active.  If `RequireAck` is configured to true