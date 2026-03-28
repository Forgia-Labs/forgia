# Forgia Dashboard

---

## Feature Designs

### In Progress / In Corso

```dataview
TABLE WITHOUT ID
  id AS "ID",
  title AS "Title",
  priority AS "P",
  status AS "Status",
  assignee AS "Assignee"
FROM "fd"
WHERE status = "in-progress" OR status = "design"
SORT priority DESC, created ASC
```

### In Review / In Revisione

```dataview
LIST WITHOUT ID id + " — " + title
FROM "fd"
WHERE status = "review" OR (reviewed = false AND status != "planned")
SORT created ASC
```

### Planned / Pianificati

```dataview
TABLE WITHOUT ID id AS "ID", title AS "Title", priority AS "P", effort AS "Effort"
FROM "fd"
WHERE status = "planned"
SORT priority DESC
```

---

## Execution Specs (SDD)

### Active / Attivi

```dataview
TABLE WITHOUT ID
  id AS "SDD",
  fd AS "FD",
  status AS "Status",
  agent AS "Agent",
  title AS "Title"
FROM "sdd"
WHERE status != "done" AND status != "failed"
SORT fd ASC, id ASC
```

### Completed / Completati

```dataview
TABLE WITHOUT ID id AS "SDD", fd AS "FD", title AS "Title"
FROM "sdd"
WHERE status = "done"
SORT file.mtime DESC
LIMIT 10
```

---

## Active Tasks / Task Attive

```dataview
TABLE WITHOUT ID
  id AS "ID",
  priority AS "P",
  owner AS "Owner",
  title AS "Task"
FROM "ops/active"
WHERE status = "active"
SORT priority ASC, due ASC
```

---

## Stats / Statistiche

```dataview
TABLE WITHOUT ID status AS "Status", length(rows) AS "Count"
FROM "fd" OR "sdd"
WHERE !contains(file.name, "_")
GROUP BY status
SORT status ASC
```
