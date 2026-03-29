package cli

import (
	"fmt"

	"connectrpc.com/connect"
	"github.com/MakeNowJust/heredoc"
	compassv1beta1 "github.com/raystack/compass/gen/raystack/compass/v1beta1"
	"github.com/raystack/compass/internal/client"
	"github.com/raystack/compass/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func entitiesCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entity",
		Aliases: []string{"entities"},
		Short:   "Manage entities in the knowledge graph",
		Annotations: map[string]string{
			"group": "core",
		},
		Example: heredoc.Doc(`
		$ compass entity list
		$ compass entity view <id>
		$ compass entity upsert
		$ compass entity delete <urn>
		$ compass entity search <text>
		$ compass entity types
		$ compass entity context <urn>
		$ compass entity impact <urn>
		`),
	}

	cmd.AddCommand(
		listEntitiesCommand(cfg),
		viewEntityCommand(cfg),
		upsertEntityCommand(cfg),
		deleteEntityCommand(cfg),
		searchEntitiesCommand(cfg),
		entityTypesCommand(cfg),
		entityContextCommand(cfg),
		entityImpactCommand(cfg),
	)

	return cmd
}

func listEntitiesCommand(cfg *config.Config) *cobra.Command {
	var types, source, query string
	var size, offset uint32

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all entities",
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.GetAllEntitiesRequest{
				Types:  types,
				Source: source,
				Q:      query,
				Size:   size,
				Offset: offset,
			})
			res, err := clnt.GetAllEntities(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println(res.Msg.GetData()); return nil
		},
	}
	cmd.Flags().StringVar(&types, "types", "", "Filter by types (comma-separated)")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Search query")
	cmd.Flags().Uint32Var(&size, "size", 20, "Page size")
	cmd.Flags().Uint32Var(&offset, "offset", 0, "Page offset")
	return cmd
}

func viewEntityCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View entity details by ID or URN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.GetEntityByIDRequest{Id: args[0]})
			res, err := clnt.GetEntityByID(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println(res.Msg.GetData()); return nil
		},
	}
}

func upsertEntityCommand(cfg *config.Config) *cobra.Command {
	var urn, typ, name, desc, source string

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update an entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.UpsertEntityRequest{
				Urn:         urn,
				Type:        typ,
				Name:        name,
				Description: desc,
				Source:      source,
				Properties:  &structpb.Struct{},
			})
			res, err := clnt.UpsertEntity(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println("Entity upserted:", res.Msg.GetId())
			return nil
		},
	}
	cmd.Flags().StringVar(&urn, "urn", "", "Entity URN (required)")
	cmd.Flags().StringVar(&typ, "type", "", "Entity type (required)")
	cmd.Flags().StringVar(&name, "name", "", "Entity name (required)")
	cmd.Flags().StringVar(&desc, "description", "", "Description")
	cmd.Flags().StringVar(&source, "source", "", "Source system")
	_ = cmd.MarkFlagRequired("urn")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func deleteEntityCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <urn>",
		Short: "Delete an entity by URN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.DeleteEntityRequest{Urn: args[0]})
			if _, err := clnt.DeleteEntity(cmd.Context(), req); err != nil {
				return err
			}

			fmt.Println("Entity deleted:", args[0])
			return nil
		},
	}
}

func searchEntitiesCommand(cfg *config.Config) *cobra.Command {
	var types, source, mode string
	var size uint32

	cmd := &cobra.Command{
		Use:   "search <text>",
		Short: "Search entities (keyword, semantic, or hybrid)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.SearchEntitiesRequest{
				Text:   args[0],
				Types:  types,
				Source: source,
				Mode:   mode,
				Size:   size,
			})
			res, err := clnt.SearchEntities(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println(res.Msg.GetData()); return nil
		},
	}
	cmd.Flags().StringVar(&types, "types", "", "Filter by types")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source")
	cmd.Flags().StringVar(&mode, "mode", "keyword", "Search mode: keyword, semantic, hybrid")
	cmd.Flags().Uint32Var(&size, "size", 10, "Max results")
	return cmd
}

func entityTypesCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List all entity types with counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.GetEntityTypesRequest{})
			res, err := clnt.GetEntityTypes(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Println(res.Msg.GetData()); return nil
		},
	}
}

func entityContextCommand(cfg *config.Config) *cobra.Command {
	var depth uint32

	cmd := &cobra.Command{
		Use:   "context <urn>",
		Short: "Get full context subgraph for an entity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.GetEntityContextRequest{
				Urn:   args[0],
				Depth: depth,
			})
			res, err := clnt.GetEntityContext(cmd.Context(), req)
			if err != nil {
				return err
			}

			fmt.Printf("Entity: %s (%s)\n", res.Msg.Entity.GetName(), res.Msg.Entity.GetType())
			if len(res.Msg.Edges) > 0 {
				fmt.Printf("\nRelationships (%d):\n", len(res.Msg.Edges))
				for _, e := range res.Msg.Edges {
					fmt.Printf("  %s —[%s]→ %s\n", e.GetSourceUrn(), e.GetType(), e.GetTargetUrn())
				}
			}
			if len(res.Msg.Related) > 0 {
				fmt.Printf("\nRelated (%d):\n", len(res.Msg.Related))
				for _, r := range res.Msg.Related {
					fmt.Printf("  %s (%s) — %s\n", r.GetName(), r.GetType(), r.GetUrn())
				}
			}
			return nil
		},
	}
	cmd.Flags().Uint32Var(&depth, "depth", 2, "Traversal depth")
	return cmd
}

func entityImpactCommand(cfg *config.Config) *cobra.Command {
	var depth uint32

	cmd := &cobra.Command{
		Use:   "impact <urn>",
		Short: "Analyze downstream blast radius",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clnt, err := createEntityClient(cmd, cfg)
			if err != nil {
				return err
			}
			

			req := connect.NewRequest(&compassv1beta1.GetEntityImpactRequest{
				Urn:   args[0],
				Depth: depth,
			})
			res, err := clnt.GetEntityImpact(cmd.Context(), req)
			if err != nil {
				return err
			}

			if len(res.Msg.Edges) == 0 {
				fmt.Println("No downstream dependencies found.")
				return nil
			}
			fmt.Printf("Impact (%d edges):\n", len(res.Msg.Edges))
			for _, e := range res.Msg.Edges {
				fmt.Printf("  %s → %s\n", e.GetSourceUrn(), e.GetTargetUrn())
			}
			return nil
		},
	}
	cmd.Flags().Uint32Var(&depth, "depth", 3, "Traversal depth")
	return cmd
}

func createEntityClient(cmd *cobra.Command, cfg *config.Config) (*client.Client, error) {
	clnt, err := client.Create(cmd.Context(), cfg.Client)
	if err != nil {
		return nil, err
	}
	return clnt, nil
}
