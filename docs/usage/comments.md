# Comments

Logged-in users can comment on any gist they can see. Comments are written in Markdown, rendered
like Markdown files, and limited to 3000 characters.

The last 5 threads are shown under the files of the gist. The Comments page of the gist, linked
from its sidebar, lists them all, oldest first and 30 threads per page.

Comments are stored in the database, not in the Git repository of the gist, so they are not part of
its revisions and are not cloned along with it.

## Replies

Any comment can be answered with a **Reply** button. Threads are one level deep: replying to a
reply adds to the same thread instead of nesting further, so the conversation stays readable on a
phone. A reply notifies the author of the comment it answers, and a thread is removed as a whole
with its first comment.

## Permissions

- The author of a comment can edit and delete it.
- The owner of a gist can delete any comment posted on it, and the whole thread below it.
- Admins can delete any comment.

An archived gist is read-only: no new comment can be posted and existing ones can't be edited, but
they can still be deleted.

Comments are deleted along with their gist, or with their author's account.

## Guest comments

The **Allow guest comments** setting lets visitors who are not signed in comment under a name they
type in. Guest comments are marked as such, have no profile page behind them, and can never be
edited or deleted by the guest who wrote them — only the gist owner and admins can delete them.

Once one third-party sign-in method is configured, people who would rather not create an account
can use it instead; guest comments are for instances that keep sign-ups closed or private.

## Notifications

When someone comments on a gist, its owner gets a notification in the inbox of the site, with a
link to the comment. When someone replies, the author of the comment being answered is notified
instead. Nobody is notified of their own comment, and the notifications can be switched off on the
whole instance in the admin panel.

## Settings

Every option lives in the **Comments** section of the admin panel:

- **Comments** — enabled by default. Turning it off hides the comment section, the Comments page of
  every gist and the comment API. Existing comments are kept and appear again once comments are
  enabled back.
- **Notify the gist owner** — enabled by default. Turning it off stops the in-site notification
  when a comment is posted.
- **Allow guest comments** — disabled by default. Turning it on opens commenting to visitors who
  are not signed in.

They can also be set before the first start with the `QYG_COMMENTS_ENABLED` environment variable;
the admin panel is what changes them afterwards. Replies and guest comments have no environment
variable: they are decided in the panel.

## API

Comments can be listed, posted, edited and deleted with the
[REST API](/docs/api/listGistComments), under `/api/gists/{uuid}/comments`. Posting, editing and
deleting comments requires a token with the `gist:write` scope. A reply is posted with a
`parent_id` field, and the list is flat: every comment carries its `parent_id` and the number of
comments on the gist is reported as `comment_count`.
