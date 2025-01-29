package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/log"
)

// AssignTags assigns a slice of tags to a given archive_id. A new tag will be implicitly created if one does not exist already. No errors will be
// given if a tag is already set. Tag aliases will automatically be resolved to their base tag.
func (a *API) AssignTags(ctx context.Context, archive_id int64, tags []string) error {
	return a.archive.AssignTags(ctx, archive_id, tags)
}

// ReplaceTags unassigns any and all tags associated with a given archive_id, and replaces it with
// a given slice of tags.
func (a *API) ReplaceTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.archive.NewSavepoint(ctx, "replacetags"); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to begin db transaction to assign tags for archive_id "+int64ToString(archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return err
	}
	defer a.archive.Rollback(ctx, "replacetags")

	a.log.LogAttrs(ctx, log.LogLevelVerbose, "removing tags for archive_id "+int64ToString(archive_id),
		slog.Int64("archive_id", archive_id),
	)
	err := a.archive.RemoveTags(ctx, archive_id)
	if err != nil {
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelVerbose, "assigning tags for archive_id "+int64ToString(archive_id),
		slog.Int64("archive_id", archive_id),
	)
	err = a.archive.AssignTags(ctx, archive_id, tags)
	if err != nil {
		return err
	}

	if err := a.archive.ReleaseSavepoint(ctx, "replacetags"); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			"failed to commit transaction for replacetags on archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}
	a.log.LogAttrs(ctx, log.LogLevelVerbose, "released savepoint 'replacetags' for archive_id "+int64ToString(archive_id),
		slog.Int64("archive_id", archive_id),
	)

	return nil
}

func (a *API) NewTagAlias(ctx context.Context, tag, tag_alias string) error {
	if tag == "" || tag_alias == "" {
		return errors.New("given empty tag or tag_alias")
	}

	if tag == tag_alias {
		return errors.New("tag is equal to tag_alias")
	}

	return a.archive.NewTagAlias(ctx, tag_alias, tag)
}

// ResolveTagAlias returns the base tag of a given tag alias.
func (a *API) ResolveTagAlias(ctx context.Context, tag_alias []string) ([]entry.TagAlias, error) {
	if tag_alias == nil {
		return nil, nil
	}

	return a.archive.ResolveTagAliasList(ctx, tag_alias)
}

// DeleteTagAlias completely deletes an alias tag from moonpool. Base tag is not affected whatsoever.
func (a *API) DeleteTagAlias(ctx context.Context, tag_alias string) error {
	if tag_alias == "" {
		return errors.New("tag_alias is empty")
	}

	return a.archive.DeleteTagAlias(ctx, tag_alias)
}

func (a *API) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	tags, err := a.archive.GetTags(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch tags for archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return nil, err
	}

	return tags, nil
}

// GetTagCountByList groups the total amount of tags that are within a range of archive_id's.
// offset is the starting point in which to begin grouping each archive_id.
// entry.TagCount is implicitly sorted from largest to smallest
func (a *API) GetTagsByRange(ctx context.Context, start, end, offset int64) ([]entry.TagCount, error) {
	return a.archive.GetTagCountByRange(ctx, start, end, end-start, offset)
}

// GetTagCountByList groups the total amount of tags that are assigned to a list of archive_id's.
// entry.TagCount is implicitly sorted from largest to smallest and returns 50 tags in total.
func (a *API) GetTagsByList(ctx context.Context, archive_ids []int64) ([]entry.TagCount, error) {
	if len(archive_ids) == 0 {
		return nil, nil
	}
	return a.archive.GetTagCountByList(ctx, archive_ids)
}

func (a *API) GetTagCount(ctx context.Context, tag string) (int64, error) {
	c, err := a.archive.GetTagCount(ctx, tag)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to get tag count for '"+tag+"'",
			slog.Any("error", err))
		return -1, err
	}

	return c, nil
}

// RemoveTags unassigns a list of tags from an entry. If a tag is no longer in reference to any entry,
// it is completely removed from the database.
func (a *API) RemoveTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.archive.NewSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			"failed to begin db transaction to remove tags for archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}
	defer a.archive.Rollback(ctx, "removetags")

	for _, tag := range tags {
		if err := a.archive.RemoveTag(ctx, archive_id, tag); err != nil {
			a.log.LogAttrs(ctx, log.LogLevelError,
				"failed to unmap tag '"+tag+"' for archive_id "+int64ToString(archive_id),
				slog.Any("error", err),
				slog.Int64("archive_id", archive_id))
			return err
		}

		t, err := a.archive.SearchTag(ctx, tag)
		if len(t) == 0 {
			if err := a.archive.DeleteTag(ctx, tag); err != nil {
				a.log.LogAttrs(ctx, log.LogLevelError, "failed to fully delete tag '"+tag+"' with no map references",
					slog.Any("error", err),
					slog.Int64("archive_id", archive_id))
				return err
			}
			a.log.LogAttrs(ctx, log.LogLevelInfo, "deleted tag '"+tag+"' due to having no more map references",
				slog.String("tag", tag),
				slog.Int64("archive_id", archive_id))
		} else {
			if err != nil {
				a.log.LogAttrs(ctx, log.LogLevelError, "failed to fully delete tag '"+tag+"' with no map references",
					slog.Any("error", err),
					slog.Int64("archive_id", archive_id))

				return err
			}
		}
	}

	if err := a.archive.ReleaseSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to commit db transaction to remove tags for archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	return nil
}

// SearchTag takes a tag and returns a slice of archive IDs.
func (a *API) SearchTag(ctx context.Context, tag string) ([]int64, error) {
	res, err := a.archive.SearchTag(ctx, tag)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to search for tag '"+tag+"'",
			slog.Any("error", err),
			slog.String("tag", tag),
		)
		return nil, err
	}

	archive_ids := make([]int64, len(res))
	for i := 0; i < len(res); i++ {
		archive_ids[i] = res[i].ID
	}

	return archive_ids, nil
}
