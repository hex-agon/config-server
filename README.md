# RuneLite Config Server

Service responsible for remote client configuration persistence.

### External Dependencies

* MongoDB
* MySQL/MariaDB

### Environment Variables

| Env                     | Description                                                                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| PORT                    | The port that the server will be bound to, defaults to 8080                                                                             |
| MONGODB_URI             | The MongoDB connection URI, more information available [here](https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection/) |
| MYSQL_URI               | The MySQL connection URI, more information available [here](https://github.com/go-sql-driver/mysql#dsn-data-source-name)                |
| MYSQL_POOL_SIZE         | The MySQL connection pool size, defaults to `10`.                                                                                       |
| MYSQL_CONN_LIFETIME     | Controls how long idle MySQL connections are kept for in minutes, defaults to `5 minutes`.                                              |
| MAX_PAYLOAD_BYTES       | The maximum acceptable payload that the server will receive in bytes, defaults to `5mb`.                                                |
| MAX_CONFIG_VALUE_LENGTH | The maximum acceptable string payload length that the server will receive, defaults to `262144`.                                        |
| NR_LICENSE              | NewRelic license key for application monitoring, if empty application monitoring will be disabled.                                      |