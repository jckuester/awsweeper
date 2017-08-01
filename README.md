# AWSweeper

Wipe your complete AWS account.

## Usage

Have a look, how to create [AWS named profiles](http://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html).
Then, to delete all resources of an account associated with a `profile`, run:

    awsweeper <profile> wipe all

To see the full list of resources which are currently supported by AWSweeper, run

     awsweeper <profile> wipe help
