# gitlab-cmds

GitLab Commands

## Getting Started

1. Install the binary

    ```
        go install github.com/jalitriver/gitlab-cmds/cmd/glcli@latest
    ```

1. Set up your authentication information as follows:

    1. Download [auth.xml.example](https://raw.githubusercontent.com/jalitriver/gitlab-cmds/master/auth.xml.example).

    1. cp auth.xml.example auth.xml
    
    1. chmod 600 auth.xml
       
    1. Edit the auth.xml file and uncomment the relevant
       authentication type and add your authentication information.

1. Set up your options file which is used to avoid having to enter the
   same command-line options whenever a command is run.

    1. Download [options.xml.example](https://raw.githubusercontent.com/jalitriver/gitlab-cmds/master/options.xml.example)
    
    1. cp options.xml.example options.xml

    1. If using a private Gitlab server, edit the options.xml file to
       point to it.

    1. You should always have an options.xml file even if everything
       (except for the root tags) is commented out; otherwise, you
       will need to pass `--options ''` to each command invocation to
       specify no options.xml file.
