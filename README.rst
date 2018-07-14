===================
ClipboardSubscriber
===================


::

  +-------+   subscribe +---------------------+  paste   +-----------+
  | Redis | <-----------| ClipboardSubscriber |--------> | Clipboard |
  +-------+             +---------------------+          +-----------+
      ^
      | publish clipboard
  +----------+
  | Somewhre |
  +----------+


Run
===

::

  $ go get github.com/pjxiao/clipboardsubscriber
  > $GOPATH\bin\clipboardsubscriber.exe --protocol=tcp --address=192.0.2.42:6379 --db=42 --subscription=clipboard
