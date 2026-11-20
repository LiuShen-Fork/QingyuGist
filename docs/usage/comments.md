# Comments

Logged-in users can comment on any gist they can see. Comments are written in Markdown, rendered
like Markdown files, and limited to 3000 characters.

The last 5 comments are shown under the files of the gist. Every comment is listed, oldest first and
30 per page, on the Comments page of the gist, linked from its sidebar.

Comments are stored in the database, not in the Git repository of the gist, so they are not part of
its revisions and are not cloned along with it.

## Permissions

- The author of a comment can edit and delete it.
- The owner of a gist can delete any comment posted on it.
- Admins can delete any comment.

An archived gist is read-only: no new comment can be posted and existing ones can't be edited, but
they can still be deleted.

Comments are deleted along with their gist, or with their author's account.

## Notifications

When someone comments on a gist, its owner gets a notification in the inbox of the site, with a
link to the comment. Owners are not notified of their own comments. The notifications can be
switched off on the whole instance in the admin panel.

## Settings

Both options live in the **Comments** section of the admin panel:

- **Comments** — enabled by default. Turning it off hides the comment section, the Comments page of
  every gist and the comment API. Existing comments are kept and appear again once comments are
  enabled back.
- **Notify the gist owner** — enabled by default. Turning it off stops the in-site notification
  when a comment is posted.

They can also be set before the first start with the `QYG_COMMENTS_ENABLED` environment variable;
the admin panel is what changes them afterwards.

## API

Comments can be listed, posted, edited and deleted with the
[REST API](/docs/api/listGistComments), under `/api/gists/{uuid}/comments`. Posting, editing and
deleting comments requires a token with the `gist:write` scope.
