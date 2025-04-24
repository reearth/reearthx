package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/reearth/reearthx/asset"
	"github.com/reearth/reearthx/asset/graph/generated"
	"github.com/reearth/reearthx/asset/graph/resolver"
	"github.com/reearth/reearthx/asset/mongo"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultPort = "8080"
const defaultMongoURI = "mongodb://localhost:27017"
const defaultMongoDatabase = "asset_example"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = defaultMongoURI
	}

	mongoDatabase := os.Getenv("MONGO_DATABASE")
	if mongoDatabase == "" {
		mongoDatabase = defaultMongoDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongodriver.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Printf("Connected to MongoDB at %s", mongoURI)

	db := mongoClient.Database(mongoDatabase)

	assetRepo := createAssetRepository(db)
	groupRepo := createGroupRepository(db)
	policyRepo := createPolicyRepository(db)
	storage, err := createStorage()
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	fileProcessor := createFileProcessor()
	zipExtractor := createZipExtractor(assetRepo, storage)

	assetService := asset.NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)
	groupService := asset.NewGroupService(groupRepo)
	policyService := asset.NewPolicyService(policyRepo)

	resolverRoot := &resolver.Resolver{
		AssetService:  assetService,
		GroupService:  groupService,
		PolicyService: policyService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolverRoot}))

	if err := os.MkdirAll("./storage", 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	http.Handle("/", playground.Handler("Asset GraphQL playground", "/query"))
	http.Handle("/query", srv)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./storage"))))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createAssetRepository(db *mongodriver.Database) asset.AssetRepository {
	return mongo.NewAssetRepository(db)
}

func createGroupRepository(db *mongodriver.Database) asset.GroupRepository {
	return mongo.NewGroupRepository(db)
}

func createPolicyRepository(db *mongodriver.Database) asset.PolicyRepository {
	return mongo.NewPolicyRepository(db)
}

func createStorage() (asset.Storage, error) {
	dir := "./storage"
	baseURL := "http://localhost:8080/assets/"
	return asset.NewLocalStorage(dir, baseURL)
}

func createFileProcessor() asset.FileProcessor {
	return asset.NewFileProcessor()
}

func createZipExtractor(assetRepo asset.AssetRepository, storage asset.Storage) asset.ZipExtractor {
	return asset.NewZipExtractor(assetRepo, storage)
}
