package source

import (
	"fmt"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/internal/log"
	"github.com/anchore/syft/syft/artifact"
)

// Location represents a path relative to a particular filesystem resolved to a specific file.Reference. This struct is used as a key
// in content fetching to uniquely identify a file relative to a request (the VirtualPath). Note that the VirtualPath
// and ref are ignored fields when using github.com/mitchellh/hashstructure. The reason for this is to ensure that
// only the minimally expressible fields of a location are baked into the uniqueness of a Location. Since VirutalPath
// and ref are not captured in JSON output they cannot be included in this minimal definition.
type Location struct {
	RealPath     string         `json:"path"`              // The path where all path ancestors have no hardlinks / symlinks
	VirtualPath  string         `hash:"ignore" json:"-"`   // The path to the file which may or may not have hardlinks / symlinks
	FileSystemID string         `json:"layerID,omitempty"` // An ID representing the filesystem. For container images this is a layer digest, directories or root filesystem this is blank.
	ref          file.Reference `hash:"ignore"`            // The file reference relative to the stereoscope.FileCatalog that has more information about this location.
}

// NewLocation creates a new Location representing a path without denoting a filesystem or FileCatalog reference.
func NewLocation(path string) Location {
	return Location{
		RealPath: path,
	}
}

// NewLocationFromImage creates a new Location representing the given path (extracted from the ref) relative to the given image.
func NewLocationFromImage(virtualPath string, ref file.Reference, img *image.Image) Location {
	entry, err := img.FileCatalog.Get(ref)
	if err != nil {
		log.Warnf("unable to find file catalog entry for ref=%+v", ref)
		return Location{
			VirtualPath: virtualPath,
			RealPath:    string(ref.RealPath),
			ref:         ref,
		}
	}

	return Location{
		VirtualPath:  virtualPath,
		RealPath:     string(ref.RealPath),
		FileSystemID: entry.Layer.Metadata.Digest,
		ref:          ref,
	}
}

// NewLocationFromDirectory creates a new Location representing the given path (extracted from the ref) relative to the given directory.
func NewLocationFromDirectory(responsePath string, ref file.Reference) Location {
	return Location{
		RealPath: responsePath,
		ref:      ref,
	}
}

func (l Location) String() string {
	str := ""
	if l.ref.ID() != 0 {
		str += fmt.Sprintf("id=%d ", l.ref.ID())
	}

	str += fmt.Sprintf("RealPath=%q", l.RealPath)

	if l.VirtualPath != "" {
		str += fmt.Sprintf(" VirtualPath=%q", l.VirtualPath)
	}

	if l.FileSystemID != "" {
		str += fmt.Sprintf(" Layer=%q", l.FileSystemID)
	}
	return fmt.Sprintf("Location<%s>", str)
}

func (l Location) ID() artifact.ID {
	f, err := artifact.DeriveID(l)
	if err != nil {
		// TODO: what to do in this case?
		log.Warnf("unable to get fingerprint of location=%+v: %+v", l, err)
		return ""
	}

	return f
}
