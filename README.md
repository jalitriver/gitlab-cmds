# gitlab-cmds

GitLab Commands

## Getting Started

1. Install the binary

    ```
        go install github.com/jalitriver/gitlab-cmds/cmd/glcmds@latest
    ```

1. Set up your options file which is used to avoid having to enter the
   same command-line options whenever a command is run.

    1. Download [options.xml.example](https://raw.githubusercontent.com/jalitriver/gitlab-cmds/master/options.xml.example)
    
    1. cp options.xml.example options.xml

    1. If using a private Gitlab server, edit the options.xml file to
       point to it.

    1. By default, glcmds looks in your current directory for
       options.xml, or you can use `glcmds --options <path>` to specify
       an alternative location.
    
    1. You should always have an options.xml file even if everything
       (except for the root tags) is commented out; otherwise, you
       will need to pass `--options ''` to each command invocation to
       specify no options.xml file.

1. Set up your authentication information as follows:

    1. Download [auth.xml.example](https://raw.githubusercontent.com/jalitriver/gitlab-cmds/master/auth.xml.example).

    1. cp auth.xml.example auth.xml
    
    1. chmod 600 auth.xml
       
    1. Edit the auth.xml file and uncomment the relevant
       authentication type and add your authentication information.
       
    1. By default, glcmds looks in your current directory for auth.xml,
       or you can use `glcmds --auth <path>` to specify an alternative
       location.  An alternative location can also be specified in the
       `options.xml` file.
       
## Managing Lists of Users

The `glcmds users list` command can be used to lookup user IDs from
usernames or e-mail addresss and output resulting list of users to
users.xml file.  The users.xml file can then be used as the input of
other commands that accept a list of users.

If users.xml does not exist, one can be created as follows:

 ```
 glcmds users list --out users.xml --users 'foo,bar'
 ```

If users.xml already exists, new users can be appended to it using the
same syntax as above :

 ```
 glcmds users list --out users.xml --users 'baz'
 ```

To remove a user from your users.xml file, just edit the file.

## Inverting --dry-run Logic

By default, all commands which can alter Gitlab will alter Gitlab
unless the `--dry-run` flag is set.  It is possible to invert this
logic so that --dry-run is enabled by default by changing the dry-run
options in options.xml to `true`.  You can then use `--dry-run=false`
on the command line when you are ready to execute the command for
real.
