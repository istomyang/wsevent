// Package wsevent provides a whole workflow from publish.Publish publishes the change events to subscribe.Subscribe.
// And then received by WS/HTTP Server, handle events and do some work, and then push new data to client.
/*
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

*/
package wsevent
