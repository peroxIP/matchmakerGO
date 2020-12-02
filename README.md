# matchmakerGO
Simple matchmaker written in GO


## Configure
`config.json` containes configuration parameters:
```json
{
    "port": 5000,
    "max_users_per_mm": 100,
    "max_users_per_party": 4
}
```

## Run the app

    go run ./

## API

### Join
`POST /join`

Request:
```json
{
    "id": 1,
}
```
Response on success:
```json 
{}
```
Response on error:
```json 
{
    "error": "msg"
}
```


### Leave
`POST /leave`

Request:
```json
{
    "id": 1,
}
```
Response on success:
```json
{}
```

Response on error:
```json 
{
    "error": "msg"
}
```


### Session
`/session`

Request: 
```json 
{}
```
Response on success (empty):
```json 
{
    "WaitingForGame": [],
    "AllSessions": {}
}
```

Response on success (populated):
```json
{
    "WaitingForGame": [
        {
            "Id": 5,
            "Status": 1,
            "SessionUUID": ""
        }
    ],
    "AllSessions": {
        "e7563576-b77b-4e2d-afa8-624a005e6527": {
            "Users": [
                {
                    "Id": 1,
                    "Status": 2,
                    "SessionUUID": "e7563576-b77b-4e2d-afa8-624a005e6527"
                },
                {
                    "Id": 2,
                    "Status": 2,
                    "SessionUUID": "e7563576-b77b-4e2d-afa8-624a005e6527"
                },
                {
                    "Id": 3,
                    "Status": 2,
                    "SessionUUID": "e7563576-b77b-4e2d-afa8-624a005e6527"
                },
                {
                    "Id": 4,
                    "Status": 2,
                    "SessionUUID": "e7563576-b77b-4e2d-afa8-624a005e6527"
                }
            ]
        }
    }
}
```

Response on error: 
```json 
{
    "error": "msg"
}
```