## Dokku Github Hook Plugin

The plugin allows github webhook deploys similar to what heroku offers. Where each commit to main can trigger a deploy to a dokku app.

### How it works
The plugin creates a webhook with github and listens for this hook in a simple http server. Users can the link the hook to a dokku app for auto deployment. The dokku app must have a repository link provided first before it can be linked to a hook.

### Install
To install the plugin, just run
``` 
dokku plugin:install https://github.com/snopan/dokku-git-hook
```

### Commands
#### Create a Webhook
Creates a webhook on the given github repo and listens for it
```
dokku github-hook:hook-create HOOK_NAME GIT_REPO_SHORT
```

The `GIT_REPO_SHORT` is in the following format `OWNER/REPO_NAME`, for further details look at the usage example below.

#### Remove a Webhook
Removes a webhook from github and stop listening for it
```
dokku github-hook:hook-remove HOOK_NAME
```

#### Provide a repository for an App
Sets a repository to use during app deployment
```
dokku github-hook:app-create APP_NAME GIT_REPO_LINK
```

#### Remove a repository for an App
Removes the deploy repository for an App
```
dokku github-hook:app-remove APP_NAME
```

### Usage example
```
dokku apps:create dokku_api_instance
dokku webhook:create api_hook github.com/user/api.git
dokku webhook:set-deploy dokku_api_instance github.com/user/api.git
dokku webhook:link dokku_api_instance api_hook

dokku apps:create dokku_ws_instance
dokku webhook:set-deploy dokku_ws_instance github.com/user/ws.git
dokku webhook:link dokku_ws_instance api_hook
```
