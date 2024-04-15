# About

This is a quick script to produce the available versions of Python in a JSON format.


# Usage

```
Usage of ./py-versions:
  -o string
        Filename to save JSON output (optional; if not provided, output to console)
  -url string
        Python FTP Mirror URL (default "https://www.python.org/ftp/python/")
```


## Example Integration

This can be combined with other scripts to alert when a new version is available. This is especially useful when working with embedded and compiled executables.

```bash
#!/bin/bash

# Define constants
ALERT_URL="https://api.example.com/v1/project-key"
TOKEN="tk_AgQdq7mVBoFD37zQVN29RhuMzNIz2"

# Log function
log() {
    echo "$(date +"%Y-%m-%d %T") $1"
}

# Get the target version from the user input, defaulting to 3.9.19 if no input is provided
if [ -z "$1" ]; then
    target_version="3.9.19"
else
    target_version="$1"
fi

# Extract the minor version from the target version
target_minor_version=$(echo "$target_version" | cut -d. -f1,2)

# Get the latest version of Python
version=$(./py-versions | grep -oP "(?<= \"$target_minor_version\":\{\"latest\":\")[^\"]+")

if [[ -z "$version" ]]; then
    log "Error: Failed to retrieve the latest version of Python."
    exit 1
fi

# Check if the latest version has a different minor version
if [[ "$version" =~ ^$target_minor_version\. ]]; then
    # Prepare alert message
    alert_message="There is a new version of Python $target_minor_version available: $version"
    
    # Send an alert with the version number using curl
    response=$(curl -s -H "Authorization: Bearer $TOKEN" -d "$alert_message" "$ALERT_URL")
    
    # Check if alert was successfully sent
    if [[ "$response" != "success" ]]; then
        log "Error: Failed to send alert. Response: $response"
        exit 1
    fi

    log "Alert sent successfully: $alert_message"
else
    log "No new version of Python $target_minor_version available."
fi
```

The above could then be combined with cron jobs to alert when a new version is available through:

```bash
0 8 * * * /path/to/example-script.sh 3.9.19
```

The `3.9.19` argument is optional and defaults to `3.9.19` if not provided. The API endpoint may trigger an action, such as sending a Slack message or initating a new build with CI/CD pipelines.
