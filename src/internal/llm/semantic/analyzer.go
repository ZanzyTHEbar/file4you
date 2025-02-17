package semantic

import "github.com/golang/geo/r3"

type FileEmbedding struct {
    Vector    r3.Vector
    Metadata  FileMetadata
}

type SemanticAnalyzer struct {
    embeddings []FileEmbedding
    threshold  float64
}

func (s *SemanticAnalyzer) FindSimilarFiles(file FileEmbedding) []FileEmbedding {
    similar := make([]FileEmbedding, 0)
    for _, e := range s.embeddings {
        similarity := cosineSimilarity(file.Vector, e.Vector)
        if similarity > s.threshold {
            similar = append(similar, e)
        }
    }
    return similar
}

func (s *SemanticAnalyzer) SuggestOrganization(file FileEmbedding) OrganizationSuggestion {
    similar := s.FindSimilarFiles(file)
    return s.analyzePatterns(similar)
}