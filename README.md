# wsevent

wsevent is an event/data flow library with enterprise level and websocket support.

```txt

         +-----------------+
         |                 |
         | Client Post/PUT |
         |                 |
         +--------+--------+
                  |
                  |
       +----------v------+----+
       |                 |    |
       | wsevent.Publish |    |
       |                 |    |
       +-----------------+    |
       |                      |
       |      API Server      |
       |                      |
       +----------+-----------+
                  |
                  | Publish Events
                  |
          +-------v--+----+
          |          |    |
          | Kafka MQ |    |
          |          |    |
          +----------+    |
          |               |
          |   Event Bus   |
          |               |
          +-------+-------+
                  |
                  | WS Push Events
                  v
       +-----------------------+
       | +-------------------+ |
       | |                   | |
       | | wsevent.Subscribe | |
       | |                   | |
       | +-------------------+ |
       | |                   | |
       | | wsevent.Dispatch  | |
       | |                   | |
       | +-------------------+ |
       | |                   | |
       | |  wsevent.Session  | |
       | |                   | |
       | --------------------+ |
       |                       |
       |     WS/HTTP Server    |
       |                       |
       +-----+-----------^-----+
             |           |
Receive Data |           | Push Message
             |           |
   +---------v-----------+---------+
   |                               |
   |         WS/HTTP Client        |
   |                               |
   +-------------------------------+

```