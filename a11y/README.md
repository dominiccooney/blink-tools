# Accessorizer

## What is it?

Accessorizer is an Android [accessibility
service](http://developer.android.com/training/accessibility/service.html). Unlike
TalkBack, Accessorizer simply dumps the accessibility events it
recieves and the accessibility tree available to it. This can be
useful for debugging accessibility problems.

You can use Accessorizer and TalkBack together, however Accessorizer
won't display the complete accessibility tree. You probably want to
use Accessorizer by itself to understand the data and events that
would be sent to TalkBack.

It is worth noting that Accessorizer is interacting with the software
under test. Both are stateful creatures. So it is certainly possible
to observe bugs with Accessorizer that TalkBack would not tickle and
vice-versa.

Accessorizer is built to debug Chromium Content Shell. You can
probably generalize it to listen to anything by hacking on the
packageNames attribute in
Accessorizer/res/xml/accessorizer_service_config.xml, but YMMV.

## How to use it

This setup is pretty fragile. Here is how to replicate it:

- Build and deploy Accessorizer.apk to a device. If you're using
  Eclipse from the Android SDK, creating a workspace pointing to a11y
  and doing File, Import, General, Existing Project into Workspace;
  then selecting Accessorizer and accepting the defaults might be
  useful.

- On your device, in Settings, Accessibility, Accessorizer, turn the
  accessibility service on.

- `adb logcat` is useful for watching Accessibility as it accepts
  connections, etc.

- Set up port forwarding so the local server can connect to the server
  built into the accessibility service: `adb forward tcp:1234
  tcp:9000`.

- Build and deploy ContentShell.apk per the [Android build
  instructions.](https://code.google.com/p/chromium/wiki/AndroidBuildInstructions) Accessorizer only listens to Content Shell.

- Install [Node](http://nodejs.org) and [Bower.](http://bower.io)

- In AccessorizerServer/static, run `bower install` to grab
  Polymer. If Bower offers you new hotness like platform#master or old
  bustedness like platform#~0.1.0, choose the new hotness.

- Per [How to Write Go Code](http://golang.org/doc/code.html) set
  GOPATH. Then `go get code.google.com/p/go.net/websocket` .

- In AccessorizerServer, `go run main.go` to start the server.

- Connect to http://localhost:9001 to load the frontend.

- Run Content Shell, for example with
  `build/android/adb_run_content_shell
  http://mattto.github.io/dialog/a11y-test.html`

- Start to tap around the UI on your device to generate accessibility
  events. The frontend you have open in your browser should light up.
