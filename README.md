# Tchoo Tchoo

The famous Tchoo Tchoo Bot for ASCR handball

## How it works ?

## Todo :
- [x] Split into services, sheet, slack, discord, ...
- [x] Slack message
- [x] Variablalize slack reciever & sheet ID
- [x] CI/CD, build & deploy with Makefile (Won't do with github action because my server is local)
- [x] Cron : tuesday & thursday, men and women
- [x] Get next match houre, team, place if we play (maybe a message when there is no match)
- [ ] Discord Log (shouterr, slack too ?)
- [ ] Rework match info to use ffhb data & a scraper (scrap every week and store in a json file)
- [ ] Documentation
- [ ] OpenTelemetry
- [ ] New commands :
  - Get last weekend stats (need to be sure that there is a least one match this weekend): score, players stats... (An AI comments about the weekend)
  - ?
- [ ] Dockerize shit (Only if web server is setup)
  - Dockerfile : ADD hello-cron /etc/cronjob; RUN crontab /etc/cronjob
