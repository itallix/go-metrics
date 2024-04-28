# go-metrics

## To get updates from template

Add git remote:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Update autotests:

```
git fetch template && git checkout template/main .github
```

## Requirements to run autotests

- Branch name should follow pattern `iter<number>`, where `<number>` — number of increment. E.g., branch `iter4` will trigger autotests for increments 1-4.
