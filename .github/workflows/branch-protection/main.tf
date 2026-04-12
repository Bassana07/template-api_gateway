provider "github" {
  token = var.github_token
  owner = var.repository_owner
}

############################
# Protection branche MAIN
############################

resource "github_branch_protection" "main" {
  repository_id = var.repository_name
  pattern       = "main"

  enforce_admins = true

  required_pull_request_reviews {
    required_approving_review_count = 1
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = false
  }

  required_status_checks {
    strict   = true
    contexts = []
  }

  allows_deletions                = false
  allows_force_pushes             = false
  require_signed_commits          = false
  require_conversation_resolution = true
}

############################
# Protection branche DEV
############################

resource "github_branch_protection" "dev" {
  repository_id = var.repository_name
  pattern       = "dev"

  enforce_admins = true

  required_pull_request_reviews {
    required_approving_review_count = 1
    dismiss_stale_reviews           = true
  }

  required_status_checks {
    strict   = true
    contexts = []
  }

  allows_deletions                = false
  allows_force_pushes             = false
  require_conversation_resolution = true
}