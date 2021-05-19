package main

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/spf13/cobra"
	"github.com/zloylos/grsync"
)

var (
	optStart = false

	restoreCmd = &cobra.Command{
		Use:   "restore <backup file>",
		Short: "restores a backup of a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("restore requires a .json or .tar backup")
			}

			if strings.HasSuffix(args[0], ".json") {
				return restore(args[0])
			} else if strings.HasSuffix(args[0], ".tar") {
				return restoreTar(args[0])
			}

			return fmt.Errorf("Unknown file type, please provide a .tar or .json file")
		},
	}
)

func restoreTar(filename string) error {
	tarfile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tr := tar.NewReader(tarfile)
	var b []byte
	for {
		th, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch th.Name {
		case "container.json":
			var err error
			b, err = ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
		}
	}

	var backup Backup
	err = json.Unmarshal(b, &backup)
	if err != nil {
		return err
	}

	id, err := createContainer(backup)
	if err != nil {
		return err
	}

	conf, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}

	tt := map[string]string{}
	for _, oldPath := range backup.Mounts {
		for _, hostPath := range conf.Mounts {
			if oldPath.Destination == hostPath.Destination {
				tt[oldPath.Source] = hostPath.Source
				break
			}
		}
	}

	if _, err := tarfile.Seek(0, 0); err != nil {
		return err
	}
	tr = tar.NewReader(tarfile)
	for {
		th, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if th.Name == "container.json" {
			continue
		}

		path := th.Name
		fmt.Println("Restoring:", path)
		for k, v := range tt {
			if strings.HasPrefix(path, k) {
				path = v + path[len(k):]
			}
		}

		if th.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(path, os.FileMode(th.Mode)); err != nil {
				return err
			}
		} else {
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tr); err != nil {
				return err
			}
			file.Close()
		}
		if err := os.Chmod(path, os.FileMode(th.Mode)); err != nil {
			return err
		}
		if err := os.Chown(path, th.Uid, th.Gid); err != nil {
			return err
		}
		fmt.Println("Created as:", path)
	}

	if optStart {
		return startContainer(id)
	}
	return nil
}

func restore(filename string) error {
	var backup Backup
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &backup)
	if err != nil {
		return err
	}

	id, err := createContainer(backup)
	if err != nil {
		return err
	}

	err = restoreFiles(filename, id, backup)
	if err != nil {
		return err
	}

	if optStart {
		return startContainer(id)
	}
	return nil
}

func createContainer(backup Backup) (string, error) {
	nameparts := strings.Split(backup.Name, "/")
	name := nameparts[len(nameparts)-1]
	fmt.Println("Restoring Container:", name)
	
	_, _, err := cli.ImageInspectWithRaw(ctx, backup.Config.Image)
	if err != nil {
		fmt.Println("Pulling Image:", backup.Config.Image)
		_, err := cli.ImagePull(ctx, backup.Config.Image, types.ImagePullOptions{})
		if err != nil {
			return "", err
		}
	}
	// io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, backup.Config, &container.HostConfig{
		PortBindings: backup.PortMap,
	}, nil, name)
	if err != nil {
		return "", err
	}
	fmt.Println("Created Container with ID:", resp.ID)

	for _, m := range backup.Mounts {
		fmt.Printf("Old Mount (type %s) %s -> %s\n", m.Type, m.Source, m.Destination)
	}

	conf, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}
	for _, m := range conf.Mounts {
		fmt.Printf("New Mount (type %s) %s -> %s\n", m.Type, m.Source, m.Destination)
	}

	return resp.ID, nil
}

func restoreFiles(filename string, id string, backup Backup) error {
	conf, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}

	tt := map[string]string{}
	for _, oldPath := range backup.Mounts {
		for _, hostPath := range conf.Mounts {
			if oldPath.Destination == hostPath.Destination {
				tt[oldPath.Source] = hostPath.Source
				break
			}
		}
	}

	outDir := strings.TrimRight(filename, ".json")
	for oldPath, newPath := range tt {
		skip := false
		for _, exclude := range optExclude {
			if strings.HasPrefix(oldPath, exclude) {
				skip = true
				break
			}
		}

		if skip {
			fmt.Println("Skipping './" + backupFiles + "'")
			continue
		}

		backupFiles := outDir + oldPath
		fmt.Println("Restoring './" + backupFiles + "' -> '" + newPath + "'")

		if info, err := os.Stat(backupFiles); err == nil && info.IsDir() {
			backupFiles += "/"
			newPath += "/"
		}

		task := grsync.NewTask(
			backupFiles,
			newPath,
			grsync.RsyncOptions{},
		)

		go func() {
			for {
				state := task.State()
				fmt.Printf(
					"\rprogress: %.2f / rem. %d / tot. %d",
					state.Progress,
					state.Remain,
					state.Total,
				)
				time.Sleep(time.Second)
			}
		}()

		err = task.Run()
		if err != nil {
			return err
		}
		fmt.Printf("\r")
	}

	return nil
}

func startContainer(id string) error {
	fmt.Println("Starting container:", id[:12])

	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	/*
		statusCh, errCh := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-statusCh:
		}

		out, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			return err
		}
		io.Copy(os.Stdout, out)
	*/

	return nil
}

func init() {
	restoreCmd.Flags().BoolVarP(&optStart, "start", "s", false, "start restored container")
	restoreCmd.Flags().StringArrayVarP(&optExclude, "exclude", "e", []string{}, "skip restoring files that start with this, can use multiple times")
	RootCmd.AddCommand(restoreCmd)
}
