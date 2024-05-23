# Registration_tg_bot

Free to use bot to provide easy registration on your events. ⭐pls

It is cross-platform, so you may can run it on any OS you want.

<!--Installation-->
## Installation

0. Go Lang installation
   
 ```rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz```
 
 ```export PATH=$PATH:/usr/local/go/bin```

1. Clone repo

```cd /usr/local/go/src```

```github.com/LLlE0/Registration_tg_bot```

2. Get into the directory

```cd Registration_tg_bot```

3. Build and run

```go build main.go```

```./main```

To set configs, you must create the '.env' file with such content:


BOT_TOKEN=<your_bot_token>

ADMIN_IDS=11111111,11111112,<etc>


<h1>LIST OF COMMANDS👾️</h1>

```/start```

- initiate registration

```/list``` 

- root only, get the list of users

```/del_{user_id}```

- root only, delete the user with ID == {user_id}

```/deleteall```

- root only, flush the table of users



