package buddyfs

// var _ server.Driver = &Driver{}

// type Driver struct {
// 	buddyClient *BuddyFsClient
// }

// func NewDriver(buddyClient *BuddyFsClient) *Driver {
// 	return &Driver{
// 		buddyClient: buddyClient,
// 	}
// }

// func (d *Driver) Stat(ctx *server.Context, path string) (os.FileInfo, error) {
// 	return d.buddyClient.Stat(path)
// }

// func (d *Driver) ListDir(ctx *server.Context, path string, fn func(os.FileInfo) error) error {

// 	file, err := d.buddyClient.Open(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	files, err := file.Readdir(100)
// 	if err != nil {
// 		return err
// 	}

// 	for _, file := range files {
// 		fn(file)
// 	}

// 	return nil

// }

// func (d *Driver) DeleteDir(ctx *server.Context, path string) error {
// 	return d.buddyClient.DeleteDir(path)
// }

// func (d *Driver) DeleteFile(ctx *server.Context, path string) error {
// 	return d.buddyClient.DeleteFile(path)
// }

/*

   // params  - a file path
   // returns - a time indicating when the requested path was last modified
   //         - an error if the file doesn't exist or the user lacks
   //           permissions
   Stat(*Context, string) (os.FileInfo, error)

   // params  - path, function on file or subdir found
   // returns - error
   //           path
   ListDir(*Context, string, func(os.FileInfo) error) error

   // params  - path
   // returns - nil if the directory was deleted or any error encountered
   DeleteDir(*Context, string) error

   // params  - path
   // returns - nil if the file was deleted or any error encountered
   DeleteFile(*Context, string) error

   // params  - from_path, to_path
   // returns - nil if the file was renamed or any error encountered
   Rename(*Context, string, string) error

   // params  - path
   // returns - nil if the new directory was created or any error encountered
   MakeDir(*Context, string) error

   // params  - path, filepos
   // returns - a string containing the file data to send to the client
   GetFile(*Context, string, int64) (int64, io.ReadCloser, error)

   // params  - destination path, an io.Reader containing the file data
   // returns - the number of bytes written and the first error encountered while writing, if any.
   PutFile(*Context, string, io.Reader, int64) (int64, error)


*/
