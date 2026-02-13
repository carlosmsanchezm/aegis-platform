package server

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	annotationAWSAccountID  = "aegis.yourorg.dev/awsAccountId"
	annotationAWSRoleARN    = "aegis.yourorg.dev/awsRoleArn"
	annotationAWSExternalID = "aegis.yourorg.dev/awsExternalId"

	envDefaultAWSAccountID  = "AEGIS_DEFAULT_AWS_ACCOUNT_ID"
	envDefaultAWSRoleARN    = "AEGIS_DEFAULT_AWS_ROLE_ARN"
	envDefaultAWSExternalID = "AEGIS_DEFAULT_AWS_EXTERNAL_ID"

	// envDevMode enables development mode which skips IAM role assumption.
	// In dev mode, AWS credentials are used directly without assuming a project role.
	// WARNING: Do not enable in production - breaks multi-tenancy isolation.
	envDevMode = "AEGIS_DEV_MODE"
)

// IsDevMode returns true if development mode is enabled.
// In dev mode, role assumption is skipped and credentials are used directly.
func IsDevMode() bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(envDevMode)))
	return val == "true" || val == "1" || val == "yes"
}

var awsAccountIDPattern = regexp.MustCompile(`^\d{12}$`)

type projectAWSCredentials struct {
	AccountID  string
	RoleARN    string
	ExternalID string
}

func (c projectAWSCredentials) toProto() *aegis.ProjectAwsCredentials {
	if !c.isComplete() {
		return nil
	}
	return &aegis.ProjectAwsCredentials{
		AccountId:  c.AccountID,
		RoleArn:    c.RoleARN,
		ExternalId: c.ExternalID,
	}
}

func (c projectAWSCredentials) isComplete() bool {
	// Account ID is always required
	if strings.TrimSpace(c.AccountID) == "" {
		return false
	}
	// In dev mode, only account ID is required
	if IsDevMode() {
		return true
	}
	// Production mode: require all fields for role assumption
	return strings.TrimSpace(c.RoleARN) != "" &&
		strings.TrimSpace(c.ExternalID) != ""
}

func (c projectAWSCredentials) validate() error {
	return validateProjectAwsCredentials(c.toProto())
}

// mergeProjectAwsDefaults fills any missing AWS fields from environment defaults.
func mergeProjectAwsDefaults(creds *aegis.ProjectAwsCredentials) *aegis.ProjectAwsCredentials {
	envAccount := strings.TrimSpace(os.Getenv(envDefaultAWSAccountID))
	envRole := strings.TrimSpace(os.Getenv(envDefaultAWSRoleARN))
	envExternal := strings.TrimSpace(os.Getenv(envDefaultAWSExternalID))

	if creds == nil {
		// Only create a creds object if any defaults are set.
		if envAccount == "" && envRole == "" && envExternal == "" {
			return nil
		}
		return &aegis.ProjectAwsCredentials{
			AccountId:  envAccount,
			RoleArn:    envRole,
			ExternalId: envExternal,
		}
	}

	// Fill missing fields from defaults.
	if strings.TrimSpace(creds.GetAccountId()) == "" && envAccount != "" {
		creds.AccountId = envAccount
	}
	if strings.TrimSpace(creds.GetRoleArn()) == "" && envRole != "" {
		creds.RoleArn = envRole
	}
	if strings.TrimSpace(creds.GetExternalId()) == "" && envExternal != "" {
		creds.ExternalId = envExternal
	}
	return creds
}

func sanitizeProjectAws(creds *aegis.ProjectAwsCredentials) *aegis.ProjectAwsCredentials {
	if creds == nil {
		return nil
	}
	sanitized := &aegis.ProjectAwsCredentials{
		AccountId:  strings.TrimSpace(creds.GetAccountId()),
		RoleArn:    strings.TrimSpace(creds.GetRoleArn()),
		ExternalId: strings.TrimSpace(creds.GetExternalId()),
	}
	if sanitized.AccountId == "" && sanitized.RoleArn == "" && sanitized.ExternalId == "" {
		return nil
	}
	return sanitized
}

func validateProjectAwsCredentials(creds *aegis.ProjectAwsCredentials) error {
	if creds == nil {
		// In dev mode, nil credentials are allowed - use ambient credentials
		if IsDevMode() {
			return nil
		}
		return fmt.Errorf("aws credentials are required")
	}

	// Account ID is always required (for resource tagging, etc.)
	if !awsAccountIDPattern.MatchString(creds.GetAccountId()) {
		return fmt.Errorf("account_id must be a 12-digit AWS account id")
	}

	// In dev mode, skip role ARN and external ID validation
	// Credentials are used directly without assuming a role
	if IsDevMode() {
		return nil
	}

	// Production mode: require role ARN and external ID for multi-tenancy
	role := creds.GetRoleArn()
	if !strings.HasPrefix(role, "arn:") || !strings.Contains(role, ":role/") {
		return fmt.Errorf("role_arn must be a valid IAM role ARN")
	}
	if creds.GetExternalId() == "" {
		return fmt.Errorf("external_id is required")
	}
	return nil
}

func populateProjectAwsFromAnnotations(p *aegis.Project) {
	if p == nil {
		return
	}
	if sanitized := sanitizeProjectAws(p.GetAws()); sanitized != nil {
		p.Aws = sanitized
		return
	}
	if creds := awsFromAnnotations(p.GetAnnotations()); creds != nil {
		p.Aws = creds
	}
}

func awsFromAnnotations(annotations map[string]string) *aegis.ProjectAwsCredentials {
	if len(annotations) == 0 {
		return nil
	}
	creds := &aegis.ProjectAwsCredentials{
		AccountId:  strings.TrimSpace(annotations[annotationAWSAccountID]),
		RoleArn:    strings.TrimSpace(annotations[annotationAWSRoleARN]),
		ExternalId: strings.TrimSpace(annotations[annotationAWSExternalID]),
	}
	if creds.AccountId == "" && creds.RoleArn == "" && creds.ExternalId == "" {
		return nil
	}
	return creds
}

func mergeProjectAnnotations(base map[string]string, aws *aegis.ProjectAwsCredentials) map[string]string {
	var merged map[string]string
	if len(base) > 0 {
		merged = make(map[string]string, len(base))
		for k, v := range base {
			key := strings.TrimSpace(k)
			if key == "" || isAWSAnnotation(key) {
				continue
			}
			value := strings.TrimSpace(v)
			if value == "" {
				continue
			}
			merged[key] = value
		}
	} else {
		merged = map[string]string{}
	}

	if aws != nil {
		merged[annotationAWSAccountID] = aws.GetAccountId()
		merged[annotationAWSRoleARN] = aws.GetRoleArn()
		merged[annotationAWSExternalID] = aws.GetExternalId()
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func isAWSAnnotation(key string) bool {
	switch key {
	case annotationAWSAccountID, annotationAWSRoleARN, annotationAWSExternalID:
		return true
	default:
		return false
	}
}

func resolveProjectCredentials(project *aegis.Project) projectAWSCredentials {
	if project == nil {
		return projectAWSCredentials{}
	}
	populateProjectAwsFromAnnotations(project)
	if project.GetAws() == nil {
		return projectAWSCredentials{}
	}
	return projectAWSCredentials{
		AccountID:  project.GetAws().GetAccountId(),
		RoleARN:    project.GetAws().GetRoleArn(),
		ExternalID: project.GetAws().GetExternalId(),
	}
}
