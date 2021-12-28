# Dokku Github Hook Plugin

The plugin allows github webhook deploys similar to what heroku offers. Where each commit to main can trigger a deploy to a dokku app. This plugin only works with repository on [Github](https://github.com).

## How it works
The plugin creates a webhook on a github repository and listens for this hook in a simple http server. Users can then link the hook to a dokku app for auto deployment. The dokku app must have a deploy repository provided first before it can be linked to a hook.

The plugin also hosts a local http server for reload calls, while it's not exposed publicaly it will still take a port.

## Getting Started
First export all the environment varibles that are required.
```
export LOCAL_CONTROL_PORT = 9000
export GITHUB_HOOK_PORT = 9090
export GITHUB_USERNAME = bob
export GITHUB_TOKEN = {github auth token}
```

Install the plugin.
``` 
dokku plugin:install https://github.com/snopan/dokku-git-hook
```

Create a hook, deploy then link.
```
dokku github-hook:create-hook api-hook bob bob-api-repo
dokku github-hook:create-deploy dokkuapp bob bob-api-repo
dokku github-hook:create-link api-hook dokkuapp
```
Now it will wait for a hook from the repo `https://github.com/bob/bob-api-repo`, and it will deploy that repo to the dokku app `dokkuapp`.

## Commands
#### `hook-create`
* Usage: `dokku github-hook:hook-create HOOK_NAME REPO_OWNER REPO_NAME`
* Description: Makes a request to github to create a webhook on the given github repo and makes this hook available for linking.
* Arguments: 
	* `HOOK_NAME` - Identifier for this new hook
	* `REPO_OWNER` - The owner of the repo to create the webhook for
	* `REPO_NAME` - The name of the repo to create the webhook for


#### `hook-remove`
* Usage: `dokku github-hook:hook-remove HOOK_NAME`
* Description: Makes a request to remove a webhook from the given github repository then remove any links that has this hook was apart of and disable it from linking.
* Arguments:
	* `HOOK_NAME` - Identifier for the hook to delete

#### `deploy-create`
* Usage: `dokku github-hook:deploy-create APP_NAME REPO_OWNER REPO_NAME`
* Description: Sets a repository for a dokku app to use during app deployment and makes this dokku app available for linking.
* Arguments:
	* `APP_NAME` - Name of the dokku app
	* `REPO_OWNER` - The owner of the repo to use for deploy
	* `REPO_NAME` - The name of the repo to use for deploy

#### `deploy-remove`
* Usage: `dokku github-hook:deploy-remove APP_NAME`
* Description: Removes the deploy repository for a dokku app and any links that this dokku app was apart of and disable it from linking.
* Arguments:
	* `APP_NAME` - Name of the dokku app

#### `link-create`
* Usage: `dokku github-hook:link-create HOOK_NAME APP_NAME`
* Description: Makes a link between a hook and a dokku app, when this hook is triggered the provided app will be deployed.
* Arguments:
	* `HOOK_NAME` - Hook identifier created in `hook-create`
	* `APP_NAME` - Dokku app that has deployment set through `deploy-create`

#### `link-remove`
* Usage: `dokku github-hook:link-remove HOOK_NAME APP_NAME`
* Description: Removes a specific link between the provided hook and dokku app.
* Arguments:
	* `HOOK_NAME` - Hook identifier that is part of a link
	* `APP_NAME` - Dokku app that is part of a link  