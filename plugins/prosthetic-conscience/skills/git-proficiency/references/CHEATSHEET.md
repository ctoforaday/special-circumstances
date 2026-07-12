# Git Surgical Recovery Cheatsheet

Tactical commands for maintaining a healthy 138GB -> 2.5MB repository state.

## 🔍 Investigation

### Find Largest Objects in History
```bash
git rev-list --objects --all | grep "$(git verify-pack -v .git/objects/pack/*.idx | sort -k 3 -n | tail -n 20 | awk '{print $1}')"
```

### Check Folder Sizes in History
```bash
git rev-list --objects --all | grep "logs/" | awk '{print $2}' | sort | uniq -c | sort -nr | head -n 20
```

## ✂️ Scrubbing (The "Nuke" Options)

### Scrub a Folder from All History
```bash
git filter-branch --force --index-filter "git rm -r --cached --ignore-unmatch <FOLDER_PATH>/" --prune-empty --tag-name-filter cat -- --all
```

### Physical Space Reclamation (Post-Scrub)
```bash
rm -rf .git/refs/original/
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

## 🩹 Recovery

### Recover a Specific File to a New Path
```bash
git show <COMMIT_HASH>:<PROJECT_PATH>/<FILE_NAME> > <LOCAL_PATH>
```

### Undo the Last Commit (Reset HEAD)
```bash
git reset <COMMIT_HASH> # Mixed reset (preserves worktree)
```

## 🧹 Housekeeping

### Find and Delete Orphaned Temp Files
```bash
find .git/objects -name ".tmp*" -delete
```
