## Dokku Webhook Deploy Plugin
### How it works
The plugin sets up a https server that listens for any webhook post request. Users can add a new hook to github and the https server. They can then link this hook to a git:sync operation which will update a given app with a given git repo.

### Instructions
Install the plugin, this will startup a https server
``` 
dokku plugin:install ...
```

Setup a webhook trigger, adds it to github and the https server
```
dokku webhook:create HOOK_NAME GIT_REPO_LINK
```

Setup what remote to pull when webhook is triggered for an app
```
dokku webhook:set-deploy APP_NAME GIT_REPO_LINK
```

Link the webhook trigger to an app that has deployment repo setup
```
dokku webhook:link APP_NAME HOOK_NAME
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
