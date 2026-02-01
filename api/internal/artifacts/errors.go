package artifacts

import (
	"errors"
	"fmt"
)

// ErrManifestIncompatible signals that a manifest is invalid or incompatible.
var ErrManifestIncompatible = errors.New("manifest incompatible")

// ErrArtifactIncompatible signals that an artifact is invalid or incompatible.
var ErrArtifactIncompatible = errors.New("artifact incompatible")

func wrapManifestError(err error) error {
	if err == nil {
		return ErrManifestIncompatible
	}
	return fmt.Errorf("%w: %v", ErrManifestIncompatible, err)
}

func wrapArtifactError(err error) error {
	if err == nil {
		return ErrArtifactIncompatible
	}
	return fmt.Errorf("%w: %v", ErrArtifactIncompatible, err)
}
