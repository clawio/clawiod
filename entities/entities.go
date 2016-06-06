package entities

const (
	// ObjectTypeTreeMimeType is the mime type assigned to tree objects when no other mime type can be inferred.
	ObjectTypeTreeMimeType string = "clawio/tree"

	// ObjectTypeBLOBMimeType is the mime type assigned to blob objects when no other mime type can be inferred.
	ObjectTypeBLOBMimeType string = "clawio/blob"
)
const (
	// ObjectTypeTree is the value assigned to objects of type tree in the "type" field.
	ObjectTypeTree ObjectType = "tree"
	// ObjectTypeBLOB is the value assigned to objects of type blob in the "type" field.
	ObjectTypeBLOB ObjectType = "blob"
)

type (
	// ObjectType indicates if the object is either a Tree or a BLOB.
	ObjectType string

	// ObjectInfo represents the metadata information retrieved
	// from an object, either tree or blob.
	ObjectInfo struct {
		PathSpec string     `json:"pathspec"`
		Size     int64      `json:"size"`
		Type     ObjectType `json:"type"`
		ModTime  int        `json:"mtime"`
		MimeType string     `json:"mime_type"`
		Checksum string     `json:"checksum"`
	}

	// User represents an user of the system.
	// They are created by the authentication service.
	User struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
)
