# Services

### [GET] /lock/\<ID\>
Creates a lock for the given id, returns `200` if successfully created, or `409` if a lock with the same id already exists

* `ID` : The identifier can be as simple as a number, or as complex as multi-path namespace, Example: `12345`, `/my/cool/lock-1234`

* `ttl` : Indicates the time in milliseconds that the lock will be active, if not entered, the lock will be active until it is removed. Example: `60`

* `wait` : Keeps connection active while lock is busy, returning when lock is released

* `lock` : Keep lock while the connection is active, as soon as the connection is closed the lock is released.

### [DELETE] /lock/\<ID\>
Releases a previously created lock, returns `200` when the lock is successfully released and `404` when the lock is not found.

* `ID` : The lock identifier to release, Example: `12345`, `/my/cool/lock-1234`
### [PATCH] /lock/\<ID\>[?ttl=\<ttl>]
Make changes to a previously created lock, return `200` when the lock is successfully changed and `404` when the lock is not found.

* `ID` : The lock identifier to release, Example: `12345`, `/my/cool/lock-1234`
* `ttl`: Changes the timeout until lock is released to the variable value, Example: `5000`
