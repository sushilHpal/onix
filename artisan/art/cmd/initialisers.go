/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-Present by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cmd

func InitialiseRootCmd() *RootCmd {
	rootCmd := NewRootCmd()
	appCmd := NewAppCmd()
	serveCmd := NewServeCmd()
	specCmd := InitialiseSpecCommand()
	buildCmd := NewBuildCmd()
	lsCmd := NewListCmd()
	pushCmd := NewPushCmd()
	rmCmd := NewRmCmd()
	tagCmd := NewTagCmd()
	runCmd := NewRunCmd()
	runCCmd := NewRunCCmd()
	mergeCmd := NewMergeCmd()
	pullCmd := NewPullCmd()
	openCmd := NewOpenCmd()
	pgpCmd := InitialisePGPCommand()
	flowCmd := InitialiseFlowCommand()
	tknCmd := InitialiseTknCommand()
	manifCmd := NewManifestCmd()
	exeCmd := NewExeCmd()
	exeCCmd := NewExeCCmd()
	waitCmd := NewWaitCmd()
	curlCmd := NewCurlCmd()
	langCmd := InitialiseLangCommand()
	envCmd := InitialiseEnvCommand()
	gitSyncCmd := NewGitSyncCmd()
	pruneCmd := NewPruneCmd()
	rootCmd.Cmd.AddCommand(
		appCmd.cmd,
		specCmd.cmd,
		serveCmd.cmd,
		buildCmd.cmd,
		lsCmd.cmd,
		pushCmd.cmd,
		rmCmd.cmd,
		tagCmd.cmd,
		runCmd.cmd,
		runCCmd.cmd,
		mergeCmd.cmd,
		pullCmd.cmd,
		openCmd.cmd,
		pgpCmd.cmd,
		flowCmd.cmd,
		manifCmd.cmd,
		exeCmd.cmd,
		exeCCmd.cmd,
		tknCmd.cmd,
		waitCmd.cmd,
		curlCmd.cmd,
		langCmd.cmd,
		envCmd.cmd,
		gitSyncCmd.cmd,
		pruneCmd.cmd,
	)
	return rootCmd
}

func InitialiseSpecCommand() *SpecCmd {
	specCmd := NewSpecCmd()
	specExportCmd := NewSpecExportCmd()
	specImportCmd := NewSpecImportCmd()
	specDownCmd := NewSpecDownCmd()
	specPushCmd := NewSpecPushCmd()
	specSignCmd := NewSpecSignCmd()
	specInfoCmd := NewSpecInfoCmd()
	specCmd.cmd.AddCommand(specExportCmd.cmd)
	specCmd.cmd.AddCommand(specImportCmd.cmd)
	specCmd.cmd.AddCommand(specDownCmd.cmd)
	specCmd.cmd.AddCommand(specPushCmd.cmd)
	specCmd.cmd.AddCommand(specSignCmd.cmd)
	specCmd.cmd.AddCommand(specInfoCmd.cmd)
	return specCmd
}

func InitialiseEnvCommand() *EnvCmd {
	envCmd := NewEnvCmd()
	envPackageCmd := NewEnvPackageCmd()
	envFlowCmd := NewEnvFlowCmd()
	envCmd.cmd.AddCommand(envFlowCmd.cmd, envPackageCmd.cmd)
	return envCmd
}

func InitialiseLangCommand() *LangCmd {
	langCmd := NewLangCmd()
	langFetchCmd := NewLangFetchCmd()
	langUpdateCmd := NewLangUpdateCmd()
	langCmd.cmd.AddCommand(langFetchCmd.cmd, langUpdateCmd.cmd)
	return langCmd
}

func InitialiseFlowCommand() *FlowCmd {
	flowCmd := NewFlowCmd()
	flowMergeCmd := NewFlowMergeCmd()
	flowRunCmd := NewFlowRunCmd()
	flowCmd.cmd.AddCommand(flowMergeCmd.cmd, flowRunCmd.cmd)
	return flowCmd
}

func InitialiseTknCommand() *TknCmd {
	tknCmd := NewTknCmd()
	tknGenCmd := NewTknGenCmd()
	tknCmd.cmd.AddCommand(tknGenCmd.cmd)
	return tknCmd
}

func InitialisePGPCommand() *PGPCmd {
	pgpCmd := NewPGPCmd()
	pgpGenCmd := NewPGPGenCmd()
	pgpImportCmd := NewPGPImportCmd()
	pgpExportCmd := NewPGPExportCmd()
	pgpEncryptCmd := NewPGPEncryptCmd()
	pgpDecryptCmd := NewPGPDecryptCmd()
	pgpCmd.cmd.AddCommand(pgpGenCmd.cmd, pgpImportCmd.cmd, pgpExportCmd.cmd, pgpEncryptCmd.cmd, pgpDecryptCmd.cmd)
	return pgpCmd
}
