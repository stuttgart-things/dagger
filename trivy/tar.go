package main

// ScanTarBallImage runs the linter on the provided source code.
// func (m *Trivy) ScanTarBallImage(
// 	ctx context.Context,
// 	file *dagger.File,
// ) (*dagger.File, error) {
// 	scans := []*dagger.TrivyScan{
// 		dag.Trivy().ImageTarball(file),
// 	}

// 	// GRAB THE REPORT AS A FILE
// 	reportFile, err := scans[0].Report("json").Sync(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting report: %w", err)
// 	}

// 	return reportFile, nil
// }
