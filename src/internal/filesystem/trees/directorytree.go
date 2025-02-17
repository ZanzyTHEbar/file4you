package trees

import (
	"context"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gonum.org/v1/gonum/spatial/kdtree"
)

type DirectoryTree struct {
	Root       *DirectoryNode
	KDTree     *kdtree.Tree             // KD-Tree structure for fast metadata-based searches
	KDTreeData DirectoryPointCollection // Holds DirectoryPoint references
	metrics    *TreeMetrics
	logger     *slog.Logger
	closeOnce  sync.Once
	cleanup    []func() error
	mu         sync.Mutex
	//Cache map[string]*DirectoryNode
}

// WithRoot sets the root directory for the DirectoryTree
func WithRoot(root string) TreeOption {
	return func(dt *DirectoryTree) {
		dt.Root = NewDirectoryNode(root, nil)
	}
}

func NewDirectoryTree(opts ...TreeOption) *DirectoryTree {
	dt := &DirectoryTree{
		metrics: &TreeMetrics{
			OperationCounts: make(map[string]int64),
			LastUpdated:     time.Now(),
		},
		logger: slog.Default(),
		//Cache: make(map[string]*DirectoryNode),
		Root: NewDirectoryNode("/", nil),
	}

	for _, opt := range opts {
		opt(dt)
	}

	return dt
}

// TreeOption allows for customization of DirectoryTree
type TreeOption func(*DirectoryTree)

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) TreeOption {
	return func(dt *DirectoryTree) {
		dt.logger = logger
	}
}

// Walk implements TreeWalker interface with context and metrics
func (dt *DirectoryTree) Walk(ctx context.Context) error {
	start := time.Now()
	defer func() {
		dt.mu.Lock()
		dt.metrics.ProcessingTime = time.Since(start)
		dt.metrics.LastUpdated = time.Now()
		dt.mu.Unlock()
	}()

	dt.logger.Info("starting tree walk",
		"root", dt.Root.Path,
		"operation", "walk",
		"timestamp", start)

	return dt.walkNode(ctx, dt.Root, 0)
}

func (dt *DirectoryTree) walkNode(ctx context.Context, node *DirectoryNode, depth int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	dt.mu.Lock()
	dt.metrics.TotalNodes++
	dt.metrics.MaxDepth = max(dt.metrics.MaxDepth, depth)
	dt.mu.Unlock()

	for _, child := range node.Children {
		if err := dt.walkNode(ctx, child, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// AddFile adds a file node to the tree at a specified path.
// If intermediate directories don't exist, it creates them.
func (tree *DirectoryTree) AddFile(path string, filePath string, size int64, modifiedAt time.Time) error {

	// Split the path into directories and then find or create the path
	targetNode := tree.FindOrCreatePath(filepath.SplitList(path))

	// Now that we're at the target directory, add the file node
	targetNode.AddFile(&FileNode{
		Path:      filePath,
		Extension: filepath.Ext(filePath),
	})

	return nil
}

// Flatten recursively collects all directories and files in a flat list of paths
func (tree *DirectoryTree) Flatten() []string {
	var paths []string
	tree.flattenNode(tree.Root, tree.Root.Path, &paths)
	return paths
}

// SafeCacheSet safely sets a value in the Cache map
//func (tree *DirectoryTree) SafeCacheSet(key string, value *DirectoryNode) {
//	tree.mu.Lock()
//	defer tree.mu.Unlock()
//
//	tree.Cache[key] = value
//}

// SafeCacheGet safely retrieves a value from the Cache map
//func (tree *DirectoryTree) SafeCacheGet(key string) (*DirectoryNode, bool) {
//	tree.mu.Lock()
//	defer tree.mu.Unlock()
//
//	value, exists := tree.Cache[key]
//	return value, exists
//}

// AddDirectory adds a directory node to the tree at a specified path
func (tree *DirectoryTree) AddDirectory(path string) (*DirectoryNode, error) {

	node := tree.Root
	segments := strings.Split(path, string(os.PathSeparator))
	for _, segment := range segments {
		found := false
		for _, child := range node.Children {
			if child.Path == segment && child.Type == Directory {
				node = child
				found = true
				break
			}
		}
		if !found {
			// Create missing directories in path
			newDir := &DirectoryNode{
				Path:     segment,
				Type:     Directory,
				Parent:   node,
				Children: []*DirectoryNode{},
				Files:    []*FileNode{},
			}
			node.Children = append(node.Children, newDir)
			node = newDir
		}
	}
	return node, nil
}

// FindOrCreatePath traverses the tree to find or create a directory path
func (tree *DirectoryTree) FindOrCreatePath(path []string) *DirectoryNode {

	current := tree.Root
	for _, dir := range path {
		var next *DirectoryNode
		for _, child := range current.Children {
			if child.Path == dir {
				next = child
				break
			}
		}
		if next == nil {
			next, _ = current.AddChildDirectory(dir)
		}
		current = next
	}
	return current
}

// Cleanup performs necessary cleanup operations on the DirectoryTree
func (dt *DirectoryTree) Cleanup() error {
	// Clear the root node
	if dt.Root != nil {
		dt.Root.Children = nil
		dt.Root.Files = nil
	}

	var err error
	dt.closeOnce.Do(func() {
		dt.logger.Info("cleaning up directory tree")
		for _, cleanup := range dt.cleanup {
			if cleanErr := cleanup(); cleanErr != nil {
				dt.logger.Error("cleanup error",
					"error", cleanErr)
				err = cleanErr
			}
		}
	})
	dt.logger.Info("directory tree cleanup complete")
	// Clear any other resources that need cleanup
	dt.Root = nil
	dt.KDTree = nil
	dt.KDTreeData = nil
	dt.metrics = nil
	dt.logger = nil
	dt.cleanup = nil
	dt.mu = sync.Mutex{}
	//dt.Cache = nil
	return err

}

// GetMetrics returns current metrics with concurrency safety
func (dt *DirectoryTree) GetMetrics(ctx context.Context) (*TreeMetrics, error) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Return a copy to prevent external modifications
	return &TreeMetrics{
		TotalNodes:      dt.metrics.TotalNodes,
		TotalSize:       dt.metrics.TotalSize,
		MaxDepth:        dt.metrics.MaxDepth,
		LastUpdated:     dt.metrics.LastUpdated,
		ProcessingTime:  dt.metrics.ProcessingTime,
		OperationCounts: maps.Clone(dt.metrics.OperationCounts),
	}, nil
}

// flattenNode is a helper function for Flatten, processing each node recursively
func (tree *DirectoryTree) flattenNode(node *DirectoryNode, currentPath string, paths *[]string) {

	// Add current directory path to paths
	*paths = append(*paths, currentPath)

	// Recursively process each child directory
	for _, child := range node.Children {
		childPath := filepath.Join(currentPath, child.Path)
		tree.flattenNode(child, childPath, paths)
	}

	// Add all files in this directory to paths
	for _, file := range node.Files {
		filePath := filepath.Join(currentPath, file.Path)
		*paths = append(*paths, filePath)
	}
}

func (tree *DirectoryTree) String() string {
	return tree.Root.String()
}

func (tree *DirectoryTree) MarshalJSON() ([]byte, error) {
	return tree.Root.MarshalJSON()
}

func (tree *DirectoryTree) UnMarshalJSON(data []byte) error {
	return tree.Root.UnMarshalJSON(data)
}
