# Serviços


### [GET] /lock/\<ID\>
Cria um lock para o id informado, retorna `200` caso seja criado com sucesso, ou `409` caso já exista um lock criado com o mesmo id

* `ID` : o identificador pode ser tão simples como um numero, ou tão complexo quando namespace com vários paths, Exemplo : `12345`, `/my/cool/lock-1234`

* `ttl` : Indica um tempo em milisegundos que o lock ficará ativo, caso não seja informado, o lock ficará ativo até que seja removido, Exemplo : `60`

* `wait`: Mantem a conexão ativa enquanto o lock estiver ocupado, retornando quando o lock for liberado

* `lock`: Mantem o lock enquanto a conexão estiver ativa, assim que a conexão for fechada o lock é liberado


### [DELETE] /lock/\<ID\>
Libera um lock criado anteriormente, retorna `200` quando o lock é liberado com sucesso e `404` quando o lock não for encontrado

* `ID` : o identificador do lock que será liberado, Exemplo : `12345`, `/my/cool/lock-1234`