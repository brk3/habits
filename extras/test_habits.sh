#!/bin/bash

BASE_URL="http://localhost:5173"

echo "🧹 Cleaning up old test data..."

# Delete old test habits
for habit in gym meditation reading running coding water yoga guitar piano journaling; do
  curl -s -X DELETE "$BASE_URL/api/habits/$habit" > /dev/null 2>&1
done

echo "✨ Creating diverse test habits..."

# Habit 1: "gym" - Epic 45-day streak (LEGENDARY)
echo "💪 gym - Epic 45-day streak (LEGENDARY)"
for i in {44..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"gym\", \"timestamp\": $DATE, \"note\": \"workout\"}" > /dev/null
done

# Habit 2: "meditation" - Strong 15-day streak
echo "🧘 meditation - Strong 15-day streak"
for i in {14..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"meditation\", \"timestamp\": $DATE, \"note\": \"session\"}" > /dev/null
done

# Habit 3: "reading" - New 4-day streak
echo "📚 reading - New 4-day streak"
for i in {3..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"reading\", \"timestamp\": $DATE, \"note\": \"chapter\"}" > /dev/null
done

# Habit 4: "running" - Broken streak (was good, then stopped)
echo "🏃 running - Broken streak (15-30 days ago)"
for i in {30..15}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"running\", \"timestamp\": $DATE, \"note\": \"5k\"}" > /dev/null
done

# Habit 5: "coding" - Perfect 60-day streak (LEGENDARY)
echo "💻 coding - Perfect 60-day streak (LEGENDARY)"
for i in {59..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"coding\", \"timestamp\": $DATE, \"note\": \"commit\"}" > /dev/null
done

# Habit 6: "water" - Just started today
echo "💧 water - Just started (1 day)"
DATE=$(date +%s)
curl -s -X POST "$BASE_URL/api/habits/" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"water\", \"timestamp\": $DATE, \"note\": \"8 glasses\"}" > /dev/null

# Habit 7: "yoga" - 8-day streak
echo "🧘 yoga - 8-day streak"
for i in {7..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"yoga\", \"timestamp\": $DATE, \"note\": \"practice\"}" > /dev/null
done

# Habit 8: "guitar" - Sporadic (no current streak)
echo "🎸 guitar - Sporadic practice (no current streak)"
for i in 45 42 38 35 30 25 20 15 10; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"guitar\", \"timestamp\": $DATE, \"note\": \"practice\"}" > /dev/null
done

# Habit 9: "piano" - Strong 20-day streak
echo "🎹 piano - Strong 20-day streak"
for i in {19..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"piano\", \"timestamp\": $DATE, \"note\": \"practice\"}" > /dev/null
done

# Habit 10: "journaling" - Perfect 90-day streak (ULTRA LEGENDARY)
echo "📝 journaling - Perfect 90-day streak (ULTRA LEGENDARY)"
for i in {89..0}; do
  DATE=$(date -v-${i}d +%s)
  curl -s -X POST "$BASE_URL/api/habits/" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"journaling\", \"timestamp\": $DATE, \"note\": \"entry\"}" > /dev/null
done

echo ""
echo "✅ Test habits created!"
echo ""
echo "Summary:"
echo "  👑 journaling: 90-day streak (LEGENDARY)"
echo "  👑 coding: 60-day streak (LEGENDARY)"
echo "  👑 gym: 45-day streak (LEGENDARY)"
echo "  🔥 piano: 20-day streak"
echo "  🔥 meditation: 15-day streak"
echo "  🔥 yoga: 8-day streak"
echo "  ⚡ reading: 4-day streak"
echo "  💧 water: 1-day streak"
echo "  ❌ running: broken streak (15-30 days ago)"
echo "  ❌ guitar: sporadic, no streak"