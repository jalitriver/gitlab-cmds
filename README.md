# gitlab-cmds

[[_TOC]]

## Getting Started

1. Build the commands you want to use as follows:

```
    go build ./cmd/<command>
```

1. Set up your authentication information as follows:

    1. Copy auth.xml.example to auth.xml.
    
    1. Change permissions on the file so only you have access to the
       file.  On Unix do the following:
       
       - chmod 600 auth.xml
       
    1. Edit auth.xml and uncomment the relevant authentication type
       and add your authentication information.
