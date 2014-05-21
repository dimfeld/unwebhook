package main

import (
	"testing"
)

func TestGithubPush(t *testing.T) {

}

func TestGitlabPush(t *testing.T) {

}

func TestGithubEvent(t *testing.T) {

}

func TestGitlabEvent(t *testing.T) {

}

const githubPush string = `{
  "ref": "refs/heads/master",
  "after": "56d108b544ffb290e2d9088bf45ff6951d4e80df",
  "before": "92a086ddb81b5c4f9438e0f01e470c73941304ad",
  "created": false,
  "deleted": false,
  "forced": false,
  "compare": "https://github.com/dimfeld/unwebhook/compare/92a086ddb81b...56d108b544ff",
  "commits": [
    {
      "id": "56d108b544ffb290e2d9088bf45ff6951d4e80df",
      "distinct": true,
      "message": "Use current directory for logs if none is given",
      "timestamp": "2014-05-20T21:27:58-05:00",
      "url": "https://github.com/dimfeld/unwebhook/commit/56d108b544ffb290e2d9088bf45ff6951d4e80df",
      "author": {
        "name": "Daniel Imfeld",
        "email": "email@example.com",
        "username": "dimfeld"
      },
      "committer": {
        "name": "Daniel Imfeld",
        "email": "email@example.com",
        "username": "dimfeld"
      },
      "added": [],
      "removed": [],
      "modified": [
        "webhook.go"
      ]
    }
  ],
  "head_commit": {
    "id": "56d108b544ffb290e2d9088bf45ff6951d4e80df",
    "distinct": true,
    "message": "Use current directory for logs if none is given",
    "timestamp": "2014-05-20T21:27:58-05:00",
    "url": "https://github.com/dimfeld/unwebhook/commit/56d108b544ffb290e2d9088bf45ff6951d4e80df",
    "author": {
      "name": "Daniel Imfeld",
      "email": "email@example.com",
      "username": "dimfeld"
    },
    "committer": {
      "name": "Daniel Imfeld",
      "email": "email@example.com",
      "username": "dimfeld"
    },
    "added": [],
    "removed": [],
    "modified": [
      "webhook.go"
    ]
  },
  "repository": {
    "id": 19957802,
    "name": "unwebhook",
    "url": "https://github.com/dimfeld/unwebhook",
    "description": "Webhook server for Gitlab and Github to run arbitrary commands based on events",
    "homepage": "",
    "watchers": 1,
    "stargazers": 1,
    "forks": 0,
    "fork": false,
    "size": 0,
    "owner": {
      "name": "dimfeld",
      "email": "daniel@danielimfeld.com"
    },
    "private": false,
    "open_issues": 0,
    "has_issues": true,
    "has_downloads": true,
    "has_wiki": true,
    "language": "Go",
    "created_at": 1400533704,
    "pushed_at": 1400639282,
    "master_branch": "master"
  },
  "pusher": {
    "name": "dimfeld",
    "email": "email@example.com"
  }
}`

const gitlabPush string = `{
  "before": "95790bf891e76fee5e1747ab589903a6a1f80f22",
  "after": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
  "ref": "refs/heads/master",
  "user_id": 4,
  "user_name": "John Smith",
  "project_id": 15,
  "repository": {
    "name": "Diaspora",
    "url": "git@example.com:diaspora.git",
    "description": "",
    "homepage": "http://example.com/diaspora"
  },
  "commits": [
    {
      "id": "b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
      "message": "Update Catalan translation to e38cb41.",
      "timestamp": "2011-12-12T14:27:31+02:00",
      "url": "http://example.com/diaspora/commits/b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
      "author": {
        "name": "Jordi Mallach",
        "email": "jordi@softcatala.org"
      }
    },
    {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "message": "fixed readme",
      "timestamp": "2012-01-03T23:36:29+02:00",
      "url": "http://example.com/diaspora/commits/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "author": {
        "name": "GitLab dev user",
        "email": "gitlabdev@dv6700.(none)"
      }
    }
  ],
  "total_commits_count": 4
}`

const gitlabIssue string = `{
  "object_kind": "issue",
  "object_attributes": {
    "id": 301,
    "title": "New API: create/update/delete file",
    "assignee_id": 51,
    "author_id": 51,
    "project_id": 14,
    "created_at": "2013-12-03T17:15:43Z",
    "updated_at": "2013-12-03T17:15:43Z",
    "position": 0,
    "branch_name": null,
    "description": "Create new API for manipulations with repository",
    "milestone_id": null,
    "state": "opened",
    "iid": 23
  }
}`
