#!/usr/bin/env bash

set -eEuxo pipefail

# Kanidm setup for Habits project
# Make sure your kanidm config file exists with:
# uri = "https://idm.aiectomy.xyz"

kanidm=(
    docker run --rm -it 
        -v ./kanidm:/root/.config/kanidm:ro
	-v ./cache:/root/.cache
        docker.io/kanidm/tools:1.7.3
        /sbin/kanidm
)

echo "Logging in as admin..."
${kanidm[@]} login -D idm_admin

echo "Creating user..."
${kanidm[@]} person create paul "Paul Bourke"

echo "Generating password reset token..."
# Generate reset token for credentials
${kanidm[@]} person credential create-reset-token paul

####
# Setup OIDC for Habits project
####

echo "Setting up groups..."
# Create habits users group
${kanidm[@]} group create habits_users

# Add user to habits_users group
${kanidm[@]} group add-members habits_users paul

echo "Creating OAuth2 application..."
# Create OAuth2 endpoint for Habits
# Update the URL to match your habits deployment
${kanidm[@]} system oauth2 create habits "Habits Tracker" https://habits.aiectomy.xyz

# Configure OAuth2 scopes
${kanidm[@]} system oauth2 update-scope-map habits habits_users openid profile

# Add redirect URL for OAuth callback
# Update this URL to match your habits app callback
${kanidm[@]} system oauth2 add-redirect-url habits https://habits.aiectomy.xyz/auth/callback

# Use short usernames for cleaner display
${kanidm[@]} system oauth2 prefer-short-username habits

echo "OAuth2 setup complete!"
echo "Here are your client credentials:"
echo "================================="

# Display client secrets (save these for your habits app config)
${kanidm[@]} system oauth2 show-basic-secret habits

echo ""
echo "Application details:"
echo "==================="
# Show full OAuth2 application details
${kanidm[@]} system oauth2 get habits

echo ""
echo "Setup complete! Next steps:"
echo "1. Use the client ID and secret in your habits app configuration"
echo "2. Set the callback URL in your habits app to match the redirect URL above"
echo "3. Configure your habits app to use these OIDC endpoints:"
echo "   - Authorization: https://idm.aiectomy.xyz:6443/ui/oauth2"
echo "   - Token: https://idm.aiectomy.xyz:6443/oauth2/token"
echo "   - UserInfo: https://idm.aiectomy.xyz:6443/oauth2/openid/paul/userinfo"
