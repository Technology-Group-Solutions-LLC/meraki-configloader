# Meraki Configloader Overview

The Meraki configloader is a tool written in golang with one specific purpose.  It will configure ports from a CSV input file, using the public meraki API.  It was created to get around limitations of the API and using postman, as JSON bodies need to be constructed without any null values.

# [Download Binaries Here Eric](https://gitlab.com/tgs-employees/meraki-configloader/-/jobs/artifacts/main/download?job=build)

## Prerequisites

Your machine needs to have golang installed to run the script.  This can be done using a package manager, or by downloading from the [GoLang website](https://golang.org/dl/).

We are also building windows executables that will simplify running the tool, the latest build from the main branch [is accessible here](https://gitlab.com/tgs-employees/meraki-configloader/-/jobs/artifacts/main/download?job=build) that includes the readme, binary, and template file.

You need to create an API key for your Meraki environmnent.  This can be done from the "My Profile" user context of the [Meraki Portal](https://account.meraki.com/secure/login/dashboard_login).

Your Meraki environment must have some level of preconfiguration before using this tool, which will most likely be done by a user interacting directly in the Meraki portal.
- Organizations must be built
- Networks must be built
- Switches must be registered and added to the appropriate network
- Switch stack relationships must be established

## Usage

Clone this repo to a working directory.

Prepare the CSV file that is provided in this repo "MerakiSwithPortCSV.csv" with your target data.  Ensure you are saving the csv file as CSV and NOT CSV UTF-8. UTF-8 has a different character for the # sign and it breaks the comment line.

```
~/p/w/meraki-configloader$  go run main.go
```

You will be prompted for input file (w/default), API target (w/default), and API key.  The tool does some basic sanity checking of input file to ensure port configs are not duplicated.  Over time / use, these sanity checks may be expanded.

```
~/p/w/meraki-configloader$  go run main.go -h
```

Adding -h will show you command line options.
Adding -debug instead will run the program in debug mode. It will only print the lines it would send to the API instead of actually sending them. This is useful for checking to make sure the CSV is being read properly and is formated correctly.
## Contributing
Merge requests are welcome; we ask that at least one peer approval is given prior to merging to the primary branch. For major changes, please open an issue first to discuss what you would like to change.
