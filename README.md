# Tchoo Tchoo

The famous Tchoo Tchoo Bot for ASCR handball

## How it works ? 

## Todo :
- [x] Split into services, sheet, slack, discord, ...
- [x] Slack message
- [ ] Discord Log (shouterr, slack too ?)
- [x] Variablalize slack reciever & sheet ID
- [ ] Cron : tuesday & thursday, men and women
  - Dockerfile : ADD hello-cron /etc/cronjob; RUN crontab /etc/cronjob 
- [ ] Dockerize shit
- [ ] Documentation
- [ ] Get next match houre, team, place if we play (maybe a message when there is no match)
- [ ] CI/CD Github action
- [ ] New commands :
  - Get last weekend stats (need to be sure that there is a least one match this weekend): score, players stats... (An AI comments about the weekend)
  - ?
