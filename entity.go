package main

type Output struct {
	EntityList       []*Entity       `json:"entityList"`
	RelationshipList []*Relationship `json:"relationshipList"`
}

type Entity struct {
	EntityName        string `json:"entity_name"`
	EntityType        string `json:"entity_type"`
	EntityDescription string `json:"entity_description"`
}

type Relationship struct {
	SourceEntity            string `json:"source_entity"`
	TargetEntity            string `json:"target_entity"`
	RelationshipDescription string `json:"relationship_description"`
	RelationshipStrength    int    `json:"relationship_strength"`
}
