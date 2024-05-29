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

```go build .```

```./Registration_tg_bot```

To set configs, you must create the '.env' file with such content:


BOT_TOKEN=<your_bot_token>

ADMIN_IDS=11111111,11111112,<etc>


<h1>List of commands👾️</h1>

```/start```

- initiate registration

```/list``` 

- admins only, get the list of users

```/del {user_id}```

- admins only, delete the user with ID == {user_id}

```/ban {user_id}```

- admins only, ban the user in the bot

```/unban {user_id}```

- admins only, unban the user in the bot

```/deleteall```

- admins only, flush the table of users and banned users

```/send | [ids] | [template]```

- admins only, do a mailing to the registered users

This function supports interpolation, so you can use a template to provide personalized mailing. Customizable fields: 
- {name}
- {email}
- {time}
- {team}

Example:

```/send | all | Hi, {name}```


```/send | 1234567890 1111111111 | Hi, user! You registered at {time} with email {email} and your team name is {team}```

##
<h2>Made by LLlE0, 11130 (c)</h2>