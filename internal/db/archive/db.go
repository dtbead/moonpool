// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package archive

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.assignTagStmt, err = db.PrepareContext(ctx, AssignTag); err != nil {
		return nil, fmt.Errorf("error preparing query AssignTag: %w", err)
	}
	if q.deleteEntryStmt, err = db.PrepareContext(ctx, DeleteEntry); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteEntry: %w", err)
	}
	if q.deleteTagStmt, err = db.PrepareContext(ctx, DeleteTag); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteTag: %w", err)
	}
	if q.deleteTagAliasStmt, err = db.PrepareContext(ctx, DeleteTagAlias); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteTagAlias: %w", err)
	}
	if q.deleteTagMapStmt, err = db.PrepareContext(ctx, DeleteTagMap); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteTagMap: %w", err)
	}
	if q.getEntryStmt, err = db.PrepareContext(ctx, GetEntry); err != nil {
		return nil, fmt.Errorf("error preparing query GetEntry: %w", err)
	}
	if q.getEntryPathStmt, err = db.PrepareContext(ctx, GetEntryPath); err != nil {
		return nil, fmt.Errorf("error preparing query GetEntryPath: %w", err)
	}
	if q.getFileMetadataStmt, err = db.PrepareContext(ctx, GetFileMetadata); err != nil {
		return nil, fmt.Errorf("error preparing query GetFileMetadata: %w", err)
	}
	if q.getHashesStmt, err = db.PrepareContext(ctx, GetHashes); err != nil {
		return nil, fmt.Errorf("error preparing query GetHashes: %w", err)
	}
	if q.getMostRecentArchiveIDStmt, err = db.PrepareContext(ctx, GetMostRecentArchiveID); err != nil {
		return nil, fmt.Errorf("error preparing query GetMostRecentArchiveID: %w", err)
	}
	if q.getMostRecentTagIDStmt, err = db.PrepareContext(ctx, GetMostRecentTagID); err != nil {
		return nil, fmt.Errorf("error preparing query GetMostRecentTagID: %w", err)
	}
	if q.getPagesByDateCreatedStmt, err = db.PrepareContext(ctx, GetPagesByDateCreated); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateCreated: %w", err)
	}
	if q.getPagesByDateCreatedDescendingStmt, err = db.PrepareContext(ctx, GetPagesByDateCreatedDescending); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateCreatedDescending: %w", err)
	}
	if q.getPagesByDateImportedAscendingStmt, err = db.PrepareContext(ctx, GetPagesByDateImportedAscending); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateImportedAscending: %w", err)
	}
	if q.getPagesByDateImportedDecendingStmt, err = db.PrepareContext(ctx, GetPagesByDateImportedDecending); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateImportedDecending: %w", err)
	}
	if q.getPagesByDateModifiedAscendingStmt, err = db.PrepareContext(ctx, GetPagesByDateModifiedAscending); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateModifiedAscending: %w", err)
	}
	if q.getPagesByDateModifiedDescendingStmt, err = db.PrepareContext(ctx, GetPagesByDateModifiedDescending); err != nil {
		return nil, fmt.Errorf("error preparing query GetPagesByDateModifiedDescending: %w", err)
	}
	if q.getPerceptualHashStmt, err = db.PrepareContext(ctx, GetPerceptualHash); err != nil {
		return nil, fmt.Errorf("error preparing query GetPerceptualHash: %w", err)
	}
	if q.getTagCountByListStmt, err = db.PrepareContext(ctx, GetTagCountByList); err != nil {
		return nil, fmt.Errorf("error preparing query GetTagCountByList: %w", err)
	}
	if q.getTagCountByRangeStmt, err = db.PrepareContext(ctx, GetTagCountByRange); err != nil {
		return nil, fmt.Errorf("error preparing query GetTagCountByRange: %w", err)
	}
	if q.getTagCountByTagStmt, err = db.PrepareContext(ctx, GetTagCountByTag); err != nil {
		return nil, fmt.Errorf("error preparing query GetTagCountByTag: %w", err)
	}
	if q.getTagIDStmt, err = db.PrepareContext(ctx, GetTagID); err != nil {
		return nil, fmt.Errorf("error preparing query GetTagID: %w", err)
	}
	if q.getTagsFromArchiveIDStmt, err = db.PrepareContext(ctx, GetTagsFromArchiveID); err != nil {
		return nil, fmt.Errorf("error preparing query GetTagsFromArchiveID: %w", err)
	}
	if q.getTimestampsStmt, err = db.PrepareContext(ctx, GetTimestamps); err != nil {
		return nil, fmt.Errorf("error preparing query GetTimestamps: %w", err)
	}
	if q.newEntryStmt, err = db.PrepareContext(ctx, NewEntry); err != nil {
		return nil, fmt.Errorf("error preparing query NewEntry: %w", err)
	}
	if q.newTagStmt, err = db.PrepareContext(ctx, NewTag); err != nil {
		return nil, fmt.Errorf("error preparing query NewTag: %w", err)
	}
	if q.newTagAliasStmt, err = db.PrepareContext(ctx, NewTagAlias); err != nil {
		return nil, fmt.Errorf("error preparing query NewTagAlias: %w", err)
	}
	if q.removeTagStmt, err = db.PrepareContext(ctx, RemoveTag); err != nil {
		return nil, fmt.Errorf("error preparing query RemoveTag: %w", err)
	}
	if q.removeTagsFromArchiveIDStmt, err = db.PrepareContext(ctx, RemoveTagsFromArchiveID); err != nil {
		return nil, fmt.Errorf("error preparing query RemoveTagsFromArchiveID: %w", err)
	}
	if q.resolveTagAliasStmt, err = db.PrepareContext(ctx, ResolveTagAlias); err != nil {
		return nil, fmt.Errorf("error preparing query ResolveTagAlias: %w", err)
	}
	if q.resolveTagAliasListStmt, err = db.PrepareContext(ctx, ResolveTagAliasList); err != nil {
		return nil, fmt.Errorf("error preparing query ResolveTagAliasList: %w", err)
	}
	if q.searchHashStmt, err = db.PrepareContext(ctx, SearchHash); err != nil {
		return nil, fmt.Errorf("error preparing query SearchHash: %w", err)
	}
	if q.searchTagStmt, err = db.PrepareContext(ctx, SearchTag); err != nil {
		return nil, fmt.Errorf("error preparing query SearchTag: %w", err)
	}
	if q.searchTagsByListDateCreatedStmt, err = db.PrepareContext(ctx, SearchTagsByListDateCreated); err != nil {
		return nil, fmt.Errorf("error preparing query SearchTagsByListDateCreated: %w", err)
	}
	if q.searchTagsByListDateImportedStmt, err = db.PrepareContext(ctx, SearchTagsByListDateImported); err != nil {
		return nil, fmt.Errorf("error preparing query SearchTagsByListDateImported: %w", err)
	}
	if q.searchTagsByListDateModifiedStmt, err = db.PrepareContext(ctx, SearchTagsByListDateModified); err != nil {
		return nil, fmt.Errorf("error preparing query SearchTagsByListDateModified: %w", err)
	}
	if q.setFileMetadataStmt, err = db.PrepareContext(ctx, SetFileMetadata); err != nil {
		return nil, fmt.Errorf("error preparing query SetFileMetadata: %w", err)
	}
	if q.setHashesStmt, err = db.PrepareContext(ctx, SetHashes); err != nil {
		return nil, fmt.Errorf("error preparing query SetHashes: %w", err)
	}
	if q.setPerceptualHashStmt, err = db.PrepareContext(ctx, SetPerceptualHash); err != nil {
		return nil, fmt.Errorf("error preparing query SetPerceptualHash: %w", err)
	}
	if q.setTimestampsStmt, err = db.PrepareContext(ctx, SetTimestamps); err != nil {
		return nil, fmt.Errorf("error preparing query SetTimestamps: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.assignTagStmt != nil {
		if cerr := q.assignTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing assignTagStmt: %w", cerr)
		}
	}
	if q.deleteEntryStmt != nil {
		if cerr := q.deleteEntryStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteEntryStmt: %w", cerr)
		}
	}
	if q.deleteTagStmt != nil {
		if cerr := q.deleteTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteTagStmt: %w", cerr)
		}
	}
	if q.deleteTagAliasStmt != nil {
		if cerr := q.deleteTagAliasStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteTagAliasStmt: %w", cerr)
		}
	}
	if q.deleteTagMapStmt != nil {
		if cerr := q.deleteTagMapStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteTagMapStmt: %w", cerr)
		}
	}
	if q.getEntryStmt != nil {
		if cerr := q.getEntryStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getEntryStmt: %w", cerr)
		}
	}
	if q.getEntryPathStmt != nil {
		if cerr := q.getEntryPathStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getEntryPathStmt: %w", cerr)
		}
	}
	if q.getFileMetadataStmt != nil {
		if cerr := q.getFileMetadataStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getFileMetadataStmt: %w", cerr)
		}
	}
	if q.getHashesStmt != nil {
		if cerr := q.getHashesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getHashesStmt: %w", cerr)
		}
	}
	if q.getMostRecentArchiveIDStmt != nil {
		if cerr := q.getMostRecentArchiveIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getMostRecentArchiveIDStmt: %w", cerr)
		}
	}
	if q.getMostRecentTagIDStmt != nil {
		if cerr := q.getMostRecentTagIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getMostRecentTagIDStmt: %w", cerr)
		}
	}
	if q.getPagesByDateCreatedStmt != nil {
		if cerr := q.getPagesByDateCreatedStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateCreatedStmt: %w", cerr)
		}
	}
	if q.getPagesByDateCreatedDescendingStmt != nil {
		if cerr := q.getPagesByDateCreatedDescendingStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateCreatedDescendingStmt: %w", cerr)
		}
	}
	if q.getPagesByDateImportedAscendingStmt != nil {
		if cerr := q.getPagesByDateImportedAscendingStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateImportedAscendingStmt: %w", cerr)
		}
	}
	if q.getPagesByDateImportedDecendingStmt != nil {
		if cerr := q.getPagesByDateImportedDecendingStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateImportedDecendingStmt: %w", cerr)
		}
	}
	if q.getPagesByDateModifiedAscendingStmt != nil {
		if cerr := q.getPagesByDateModifiedAscendingStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateModifiedAscendingStmt: %w", cerr)
		}
	}
	if q.getPagesByDateModifiedDescendingStmt != nil {
		if cerr := q.getPagesByDateModifiedDescendingStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPagesByDateModifiedDescendingStmt: %w", cerr)
		}
	}
	if q.getPerceptualHashStmt != nil {
		if cerr := q.getPerceptualHashStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPerceptualHashStmt: %w", cerr)
		}
	}
	if q.getTagCountByListStmt != nil {
		if cerr := q.getTagCountByListStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTagCountByListStmt: %w", cerr)
		}
	}
	if q.getTagCountByRangeStmt != nil {
		if cerr := q.getTagCountByRangeStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTagCountByRangeStmt: %w", cerr)
		}
	}
	if q.getTagCountByTagStmt != nil {
		if cerr := q.getTagCountByTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTagCountByTagStmt: %w", cerr)
		}
	}
	if q.getTagIDStmt != nil {
		if cerr := q.getTagIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTagIDStmt: %w", cerr)
		}
	}
	if q.getTagsFromArchiveIDStmt != nil {
		if cerr := q.getTagsFromArchiveIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTagsFromArchiveIDStmt: %w", cerr)
		}
	}
	if q.getTimestampsStmt != nil {
		if cerr := q.getTimestampsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getTimestampsStmt: %w", cerr)
		}
	}
	if q.newEntryStmt != nil {
		if cerr := q.newEntryStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing newEntryStmt: %w", cerr)
		}
	}
	if q.newTagStmt != nil {
		if cerr := q.newTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing newTagStmt: %w", cerr)
		}
	}
	if q.newTagAliasStmt != nil {
		if cerr := q.newTagAliasStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing newTagAliasStmt: %w", cerr)
		}
	}
	if q.removeTagStmt != nil {
		if cerr := q.removeTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing removeTagStmt: %w", cerr)
		}
	}
	if q.removeTagsFromArchiveIDStmt != nil {
		if cerr := q.removeTagsFromArchiveIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing removeTagsFromArchiveIDStmt: %w", cerr)
		}
	}
	if q.resolveTagAliasStmt != nil {
		if cerr := q.resolveTagAliasStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing resolveTagAliasStmt: %w", cerr)
		}
	}
	if q.resolveTagAliasListStmt != nil {
		if cerr := q.resolveTagAliasListStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing resolveTagAliasListStmt: %w", cerr)
		}
	}
	if q.searchHashStmt != nil {
		if cerr := q.searchHashStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchHashStmt: %w", cerr)
		}
	}
	if q.searchTagStmt != nil {
		if cerr := q.searchTagStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchTagStmt: %w", cerr)
		}
	}
	if q.searchTagsByListDateCreatedStmt != nil {
		if cerr := q.searchTagsByListDateCreatedStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchTagsByListDateCreatedStmt: %w", cerr)
		}
	}
	if q.searchTagsByListDateImportedStmt != nil {
		if cerr := q.searchTagsByListDateImportedStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchTagsByListDateImportedStmt: %w", cerr)
		}
	}
	if q.searchTagsByListDateModifiedStmt != nil {
		if cerr := q.searchTagsByListDateModifiedStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing searchTagsByListDateModifiedStmt: %w", cerr)
		}
	}
	if q.setFileMetadataStmt != nil {
		if cerr := q.setFileMetadataStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing setFileMetadataStmt: %w", cerr)
		}
	}
	if q.setHashesStmt != nil {
		if cerr := q.setHashesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing setHashesStmt: %w", cerr)
		}
	}
	if q.setPerceptualHashStmt != nil {
		if cerr := q.setPerceptualHashStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing setPerceptualHashStmt: %w", cerr)
		}
	}
	if q.setTimestampsStmt != nil {
		if cerr := q.setTimestampsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing setTimestampsStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                                   DBTX
	tx                                   *sql.Tx
	assignTagStmt                        *sql.Stmt
	deleteEntryStmt                      *sql.Stmt
	deleteTagStmt                        *sql.Stmt
	deleteTagAliasStmt                   *sql.Stmt
	deleteTagMapStmt                     *sql.Stmt
	getEntryStmt                         *sql.Stmt
	getEntryPathStmt                     *sql.Stmt
	getFileMetadataStmt                  *sql.Stmt
	getHashesStmt                        *sql.Stmt
	getMostRecentArchiveIDStmt           *sql.Stmt
	getMostRecentTagIDStmt               *sql.Stmt
	getPagesByDateCreatedStmt            *sql.Stmt
	getPagesByDateCreatedDescendingStmt  *sql.Stmt
	getPagesByDateImportedAscendingStmt  *sql.Stmt
	getPagesByDateImportedDecendingStmt  *sql.Stmt
	getPagesByDateModifiedAscendingStmt  *sql.Stmt
	getPagesByDateModifiedDescendingStmt *sql.Stmt
	getPerceptualHashStmt                *sql.Stmt
	getTagCountByListStmt                *sql.Stmt
	getTagCountByRangeStmt               *sql.Stmt
	getTagCountByTagStmt                 *sql.Stmt
	getTagIDStmt                         *sql.Stmt
	getTagsFromArchiveIDStmt             *sql.Stmt
	getTimestampsStmt                    *sql.Stmt
	newEntryStmt                         *sql.Stmt
	newTagStmt                           *sql.Stmt
	newTagAliasStmt                      *sql.Stmt
	removeTagStmt                        *sql.Stmt
	removeTagsFromArchiveIDStmt          *sql.Stmt
	resolveTagAliasStmt                  *sql.Stmt
	resolveTagAliasListStmt              *sql.Stmt
	searchHashStmt                       *sql.Stmt
	searchTagStmt                        *sql.Stmt
	searchTagsByListDateCreatedStmt      *sql.Stmt
	searchTagsByListDateImportedStmt     *sql.Stmt
	searchTagsByListDateModifiedStmt     *sql.Stmt
	setFileMetadataStmt                  *sql.Stmt
	setHashesStmt                        *sql.Stmt
	setPerceptualHashStmt                *sql.Stmt
	setTimestampsStmt                    *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                                   tx,
		tx:                                   tx,
		assignTagStmt:                        q.assignTagStmt,
		deleteEntryStmt:                      q.deleteEntryStmt,
		deleteTagStmt:                        q.deleteTagStmt,
		deleteTagAliasStmt:                   q.deleteTagAliasStmt,
		deleteTagMapStmt:                     q.deleteTagMapStmt,
		getEntryStmt:                         q.getEntryStmt,
		getEntryPathStmt:                     q.getEntryPathStmt,
		getFileMetadataStmt:                  q.getFileMetadataStmt,
		getHashesStmt:                        q.getHashesStmt,
		getMostRecentArchiveIDStmt:           q.getMostRecentArchiveIDStmt,
		getMostRecentTagIDStmt:               q.getMostRecentTagIDStmt,
		getPagesByDateCreatedStmt:            q.getPagesByDateCreatedStmt,
		getPagesByDateCreatedDescendingStmt:  q.getPagesByDateCreatedDescendingStmt,
		getPagesByDateImportedAscendingStmt:  q.getPagesByDateImportedAscendingStmt,
		getPagesByDateImportedDecendingStmt:  q.getPagesByDateImportedDecendingStmt,
		getPagesByDateModifiedAscendingStmt:  q.getPagesByDateModifiedAscendingStmt,
		getPagesByDateModifiedDescendingStmt: q.getPagesByDateModifiedDescendingStmt,
		getPerceptualHashStmt:                q.getPerceptualHashStmt,
		getTagCountByListStmt:                q.getTagCountByListStmt,
		getTagCountByRangeStmt:               q.getTagCountByRangeStmt,
		getTagCountByTagStmt:                 q.getTagCountByTagStmt,
		getTagIDStmt:                         q.getTagIDStmt,
		getTagsFromArchiveIDStmt:             q.getTagsFromArchiveIDStmt,
		getTimestampsStmt:                    q.getTimestampsStmt,
		newEntryStmt:                         q.newEntryStmt,
		newTagStmt:                           q.newTagStmt,
		newTagAliasStmt:                      q.newTagAliasStmt,
		removeTagStmt:                        q.removeTagStmt,
		removeTagsFromArchiveIDStmt:          q.removeTagsFromArchiveIDStmt,
		resolveTagAliasStmt:                  q.resolveTagAliasStmt,
		resolveTagAliasListStmt:              q.resolveTagAliasListStmt,
		searchHashStmt:                       q.searchHashStmt,
		searchTagStmt:                        q.searchTagStmt,
		searchTagsByListDateCreatedStmt:      q.searchTagsByListDateCreatedStmt,
		searchTagsByListDateImportedStmt:     q.searchTagsByListDateImportedStmt,
		searchTagsByListDateModifiedStmt:     q.searchTagsByListDateModifiedStmt,
		setFileMetadataStmt:                  q.setFileMetadataStmt,
		setHashesStmt:                        q.setHashesStmt,
		setPerceptualHashStmt:                q.setPerceptualHashStmt,
		setTimestampsStmt:                    q.setTimestampsStmt,
	}
}
