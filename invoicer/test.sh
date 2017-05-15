#! /bin/bash

echo "opening profiles"
invoicer tx invoicer profile-open Frey --from key.json --amount 1mycoin
invoicer tx invoicer profile-open Rige --from key.json --amount 1mycoin

echo "sending a wage invoice"
invoicer tx invoicer wage-open Rige 20.1BTC --to Frey --notes wudduxxp --from key.json --amount 1mycoin

echo "query all invoices"
echo "invoicer query invoices"
invoicer query invoices

echo "closing the opened invoice"
echo "invoicer tx invoicer close-invoice 0x0170C7DE3179AFF6D1AFF4F83FD02896B9DCA8B130 --cur 10BTC --id "Tranzact10" --from key.json --amount 1mycoiu"
invoicer tx invoicer close-invoice 0x0170C7DE3179AFF6D1AFF4F83FD02896B9DCA8B130 --cur 10BTC --id "Tranzact10" --from key.json --amount 1mycoin

echo "query all invoices"
invoicer query invoices

