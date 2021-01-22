package soundcloader

import (
	"fmt"
	"os"
	"os/exec"
)

//func ffmpegGet(uri string, outName string) error {
//	args := make([]string, 0)
//	args = append(args,
//		"-y",             // overwrite existing
//		"-i", uri, // input file
//		"-vn",
//		"-ar", "44100",
//		"-ac", "2",
//		"-b:a", "128k")
//	args = append(args, outName) // output name should always be latest element
//	// if need debug
//	//args = append(args, "-report")
//	_, err := execute(ffmpegFindBin(), args...)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (c *Client) ffmpegGet(uri string, outName string) error {
	args := make([]string, 0)
	args = append(args,
		"-hide_banner",
		"-i", uri,
		"-c", "copy",
	)
	args = append(args, outName)

	if c.Debug {
		args = append(args, "-report")
	}

	res, err := execute(ffmpegFindBin(), args...)
	if err != nil {
		return fmt.Errorf("[FFMPEG] %s: %s", err, res)
	}
	return nil
}

func (c *Client) ffmpegUpdateTags(filename string, metadata []string) error {
	tmpFileName := fmt.Sprintf("%s.tmp", filename)
	if err := os.Rename(filename, tmpFileName); err != nil {
		return err
	}
	defer os.Remove(tmpFileName)

	args := make([]string, 0)
	args = append(args,
		"-y",
		"-hide_banner",
		"-i", tmpFileName,
		"-c", "copy")
	for _, el := range metadata {
		args = append(args, "-metadata", el)
	}
	args = append(args, filename) // output name should always be latest element

	if c.Debug {
		args = append(args, "-report")
	}

	res, err := execute(ffmpegFindBin(), args...)
	if err != nil {
		return fmt.Errorf("[FFMPEG] %s: %s", err, res)
	}

	return nil
}
func (c *Client) ffmpegAddThumbnail(filename string, thumb string) error {
	tmpFileName := fmt.Sprintf("%s.tmp", filename)
	if err := os.Rename(filename, tmpFileName); err != nil {
		return err
	}
	defer os.Remove(tmpFileName)

	thumbnailLocation, err := c.downloadThumbnail(thumb)
	if err != nil {
		return err
	}
	defer os.Remove(thumbnailLocation)

	args := make([]string, 0)
	args = append(args,
		"-y", // overwrite existing
		"-hide_banner",
		"-i", tmpFileName, // input file
		"-i", thumbnailLocation,
		//"-vn",
		"-map", "0:0", "-map", "1:0",
		"-c", "copy",
		"-id3v2_version", "3",
		"-metadata:s:v", "title=Album cover",
		"-metadata:s:v", "comment=Cover (front)",
		filename)

	if c.Debug {
		args = append(args, "-report")
	}

	_, err = execute(ffmpegFindBin(), args...)
	if err != nil {
		return err
	}

	return nil
}

func ffmpegFindBin() string {
	bin := "ffmpeg"
	path, err := exec.LookPath(bin)
	if err != nil {
		path = bin
	}
	return path
}

func execute(app string, args ...string) ([]byte, error) {
	cmd := exec.Command(app, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}
