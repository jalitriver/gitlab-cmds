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
names, usernames, or e-mail addresses and output the resulting list of
users to users.xml file.  The users.xml file can then be used as the
input of other commands that accept a list of users.

At the time of writing, searching by e-mail address does not seem to
work.  I am not sure if that is a privacy feature that need to be
disabled server-side or not.  At any rate, if you are not sure of the
username, you can do a substring search which will list all usernames
and names that contain the substring:

 ```
 glcmds users list --match-substrings --users foo
 ```

If users.xml does not exist, one can be created as follows:

 ```
 glcmds users list --out users.xml --users 'foobar'
 ```

If users.xml already exists, new users can be appended to it using the
same syntax as above :

 ```
 glcmds users list --out users.xml --users 'baz'
 ```

If you already have list of user ID's they can also be added to the
`--users` flag as follows:

 ```
 glcmds users list --out users.xml --users '1,2,3'
 ```

To remove a user from your users.xml file, just edit the file.

## Batch Approval Rule Updates for List of Approvers

To update the approvers for approval rules, you must first create an
XML file that lists the approvers.  You do this using the `glcmds
users` command as explained in the "Managing Lists of Users" section
above.  The remainder of this section assumes you have done this and
saved the results to `users.xml`.

After setting up your list of approvers, to recursively update all
approval rules for projects under a particular group, do the following
(with the `--dry-run` option) just to see what would be updated:

 ```
 glcmds projects approval-rules update --recursive --group <group> --approvers users.xml --dry-run
 ```

If everything looks correct, do the following (without the `--dry-run`
option) to perform the updated:

 ```
 glcmds projects approval-rules update --recursive --group <group> --approvers users.xml
 ```
 
 If you want to limit the projects for which approval rules are
 updated, use the `--expr` option to provide a regular expression
 which will select the projects for which approval rules will be
 updated.  The command below should be run first with and then without
 the `--dry-run` option:

 ```
 glcmds projects approval-rules update --recursive --group <group> --approvers users.xml --expr 'foo/bar/baz'
 ```
 
## Inverting --dry-run Logic

By default, all commands which can alter Gitlab will alter Gitlab
unless the `--dry-run` flag is set.  It is possible to invert this
logic so that --dry-run is enabled by default by changing the dry-run
options in options.xml to `true`.  You can then use `--dry-run=false`
on the command line when you are ready to execute the command for
real.
