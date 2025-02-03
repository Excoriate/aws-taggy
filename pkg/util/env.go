package util

import (
	"fmt"
	"os"
)

// ScanAWSEnvVars scans and retrieves all environment variables that start with the "AWS_" prefix.
//
// This function searches through all current environment variables and collects those
// that begin with "AWS_" into a map. It provides a convenient way to extract AWS-related
// configuration from the environment.
//
// Returns:
//   - A map of AWS environment variables where the key is the full variable name
//   - An error if no AWS-related environment variables are found
func ScanAWSEnvVars() (map[string]string, error) {
	awsVars := make(map[string]string)

	// Specific AWS-related environment variables to look for
	awsSpecificVars := map[string]bool{
		"AWS_REGION":            true,
		"AWS_DEFAULT_REGION":    true,
		"AWS_ACCESS_KEY_ID":     true,
		"AWS_SECRET_ACCESS_KEY": true,
		"AWS_SESSION_TOKEN":     true,
	}

	// Iterate through all environment variables
	for _, env := range os.Environ() {
		// Split the environment variable into name and value
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				name := env[:i]
				value := env[i+1:]

				// Check if the variable is in our specific AWS variables list
				if awsSpecificVars[name] && value != "" {
					awsVars[name] = value
				}
			}
		}
	}

	// Return an error if no AWS variables were found
	if len(awsVars) == 0 {
		return nil, fmt.Errorf("no AWS environment variables found")
	}

	return awsVars, nil
}

func GetAWSRegionEnvVar() (string, error) {
	awsVars, err := ScanAWSEnvVars()
	if err != nil {
		return "", fmt.Errorf("failed to scan AWS environment variables: %w", err)
	}

	return awsVars["AWS_REGION"], nil
}

func GetAWSRegionDefaultEnvVar() (string, error) {
	awsVars, err := ScanAWSEnvVars()
	if err != nil {
		return "", fmt.Errorf("failed to scan AWS environment variables: %w", err)
	}

	return awsVars["AWS_DEFAULT_REGION"], nil
}

func GetAWSAccessKeyIDEnvVar() (string, error) {
	awsVars, err := ScanAWSEnvVars()
	if err != nil {
		return "", fmt.Errorf("failed to scan AWS environment variables: %w", err)
	}

	return awsVars["AWS_ACCESS_KEY_ID"], nil
}

func GetAWSSecretAccessKeyEnvVar() (string, error) {
	awsVars, err := ScanAWSEnvVars()
	if err != nil {
		return "", fmt.Errorf("failed to scan AWS environment variables: %w", err)
	}

	return awsVars["AWS_SECRET_ACCESS_KEY"], nil
}
