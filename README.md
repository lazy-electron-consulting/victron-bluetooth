# victron-bluetooth

Reads device data via Bluetooth Low Energy (BLE) product advertisements.
Currently only supports battery monitors. Intended to be used as a library or as binaries.

## Requirements (rough)

1. use `bluetoothctl` to pair ahead of time
2. enable the GATT protocol in the victron app
3. jump through a lot of hoops to get the decryption key, see <https://github.com/keshavdv/victron-ble>

## References

- <https://github.com/keshavdv/victron-ble>
- <https://github.com/birdie1/victron>
- <https://community.victronenergy.com/questions/93919/victron-bluetooth-ble-protocol-publication.html>
- <https://community.victronenergy.com/questions/187303/victron-bluetooth-advertising-protocol.html>
- <https://community.victronenergy.com/storage/attachments/48745-extra-manufacturer-data-2022-12-14.pdf>
- <https://community.victronenergy.com/questions/43667/victronconnect-for-linux-download-instructions.html>
