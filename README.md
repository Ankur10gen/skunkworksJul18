**This code repository was created during a PoC and should not be followed for ideas on standard practice**

1. Test cases were created earlier during unit tests and the code was modified later during integrations
1. Some of the code snippets have been copied from internet for quick workarounds. I have tried to mention sources in the blog but I'm sure I missed a few 
1. S3 credentials are being picked from `~/.aws/credentials` file on the host running the application
1. Simple `log.Fatal()` has been used to handle errors which is not recommended