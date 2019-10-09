# Services

### [GET] /lock/\<ID\>

Creates a lock for the given id, returns `200` if successfully created,or `409` if a lock with the same id already exists

- `ID` : The identifier may be as simple as a number or as complex as a namespace with multiple paths. Example: `12345`, `/my/cool/lock-1234`

- `ttl` : Indicates a time in milliseconds that the lock will be active, if not entered, the lock will be active until it is removed. Example: `60`

- `wait`: Keeps connection active while lock is busy, returning when lock is released.

- `lock`: Keeps lock while the connection is active. As soon as the connection is closed the lock is released.

### [DELETE] /lock/\<ID\>

Releases a previously created lock. Returns `200` when the lock is successfully released and `404` when the lock is not found.

- `ID` : The lock identifier that will be released, Example: `12345`, `/my/cool/lock-1234`

### [PATCH] /lock/\<ID\>[?ttl=\<ttl>]

Makes changes to a previously created lock. Returns `200` when lock is successfully changed and `404` lock not found

- `ID` : The lock identifier that will be released, Example : `12345`, `/my/cool/lock-1234`
- `ttl`: Changes the timeout until lock is released to the variable value, Example : `5000`
