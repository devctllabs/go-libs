// Package filesystem provides rooted operating system filesystem operations
// that compose with the standard io/fs package.
//
// # Integration
//
// Read sources use fs.FS directly, so callers can supply embed.FS, os.DirFS,
// fstest.MapFS, or another implementation without an adapter. Open returns an
// OS rooted at an existing operating system directory. OS implements fs.FS,
// fs.ReadLinkFS, Copier, Merger, Writer, Remover, and io.Closer. It also
// publishes regular files and complete directory snapshots through its
// PublishFile and PublishDirectory methods.
//
// Infrastructure components that directly coordinate filesystem mechanics may
// accept only the narrow interface they use, such as Copier or Writer, and
// receive *OS from their application composition root. Service and usecase
// packages should usually depend on consumer-owned application capabilities
// such as PackageStore or Workspace instead of exposing generic filesystem
// operations, paths, layout, or symlink policy as business contracts.
//
// Do not inject a filesystem abstraction merely to avoid using os in a concrete
// filesystem repository. Test that repository against t.TempDir unless it has
// a real alternate backend or coordinates behavior that benefits from a lower
// seam. Generated gomock implementations of this package's capability
// interfaces are available in the filesystem/mocks subpackage.
//
// # Root and operation semantics
//
// The caller must close an OS. All names used after Open must satisfy
// fs.ValidPath and cannot escape the root through parent traversal or symbolic
// links. Mutating operations honor a canceled context before changing the
// destination; Copy and Merge also check it while processing a tree.
//
// Copy preserves existing files, while Merge replaces regular files and
// symbolic links only when their destination types match. Both operations
// preserve unrelated destination entries, copy symbolic links as links, and
// may leave a partial result if they fail or their context is canceled. They
// do not provide atomic replacement or rollback. New regular files follow
// os.CopyFS permission semantics; replacing an existing regular file preserves
// its destination permissions.
//
// PublishFile prepares a sibling temporary file before replacing its target.
// PublishDirectory prepares a complete sibling tree before replacing its
// target with a backup-and-rename sequence. The latter prevents a partial tree
// from being published on platforms with atomic sibling renames, but it can
// briefly leave the target absent and does not claim power-loss durability.
package filesystem
