package azuretpl

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

var (
	cicdMaskVarNumber atomic.Uint64
)

func (e *AzureTemplateExecutor) handleCicdMaskSecret(val string) {
	workflowLogMsgList := []string{}

	if !e.opts.Debug && !e.opts.Stdout {
		// no mask hiding needed
		return
	}

	// only show first line of error (could be a multi line error message)
	val = strings.SplitN(val, "\n", 2)[0]

	switch {
	case os.Getenv("SYSTEM_TEAMFOUNDATIONSERVERURI") != "":
		// Azure DevOps
		workflowLogMsgList = append(
			workflowLogMsgList,
			fmt.Sprintf(`##vso[task.setvariable variable=HELM_AZURETPL_SECRET_MASK_%d;isSecret=true]%v`, cicdMaskVarNumber.Add(1), val),
			fmt.Sprintf(`##vso[task.setvariable variable=HELM_AZURETPL_SECRET_MASK_%d;isSecret=true]%v`, cicdMaskVarNumber.Add(1), base64.StdEncoding.EncodeToString([]byte(val))),
		)
	case os.Getenv("GITLAB_CI") != "":
		// GitLab
		// no secret masking available
	case os.Getenv("JENKINS_URL") != "":
		// Jenkins
		// no secret masking available
	case os.Getenv("GITHUB_ACTION") != "":
		// GitHub
		workflowLogMsgList = append(
			workflowLogMsgList,
			fmt.Sprintf(`::add-mask::%v`, val),
			fmt.Sprintf(`::add-mask::%v`, base64.StdEncoding.EncodeToString([]byte(val))),
		)
	}

	for _, workflowLogMsg := range workflowLogMsgList {
		e.logger.Sync() //nolint:errcheck
		fmt.Fprintln(os.Stderr, workflowLogMsg)
		e.logger.Sync() //nolint:errcheck
	}
}

func (e *AzureTemplateExecutor) handleCicdWarning(err error) error {
	workflowLogMsg := ""

	// only show first line of error (could be a multi line error message)
	workflowLogError := strings.SplitN(err.Error(), "\n", 2)[0]

	switch {
	case os.Getenv("SYSTEM_TEAMFOUNDATIONSERVERURI") != "":
		// Azure DevOps
		workflowLogMsg = fmt.Sprintf(`##vso[task.logissue type=warning;sourcepath=%v]%v`, e.currentPath, workflowLogError)
	case os.Getenv("GITLAB_CI") != "":
		// GitLab
		// no error logging available
	case os.Getenv("JENKINS_URL") != "":
		// Jenkins
		// no error logging available
	case os.Getenv("GITHUB_ACTION") != "":
		// GitHub
		workflowLogMsg = fmt.Sprintf(`::warning file=%v,title=helm-azure-tpl::%v`, e.currentPath, workflowLogError)
	}

	if workflowLogMsg != "" {
		e.logger.Sync() //nolint:errcheck
		fmt.Fprintln(os.Stderr, workflowLogMsg)
		e.logger.Sync() //nolint:errcheck
	}

	return err
}

func (e *AzureTemplateExecutor) handleCicdError(err error) error {
	workflowLogMsg := ""

	// only show first line of error (could be a multi line error message)
	workflowLogError := strings.SplitN(err.Error(), "\n", 2)[0]

	switch {
	case os.Getenv("SYSTEM_TEAMFOUNDATIONSERVERURI") != "":
		// Azure DevOps
		workflowLogMsg = fmt.Sprintf(`##vso[task.logissue type=error;sourcepath=%v]%v`, e.currentPath, workflowLogError)
	case os.Getenv("GITLAB_CI") != "":
		// GitLab
		// no error logging available
	case os.Getenv("JENKINS_URL") != "":
		// Jenkins
		// no error logging available
	case os.Getenv("GITHUB_ACTION") != "":
		// GitHub
		workflowLogMsg = fmt.Sprintf(`::error file=%v,title=helm-azure-tpl::%v`, e.currentPath, workflowLogError)
	}

	if workflowLogMsg != "" {
		e.logger.Sync() //nolint:errcheck
		fmt.Fprintln(os.Stderr, workflowLogMsg)
		e.logger.Sync() //nolint:errcheck
	}

	return err
}
