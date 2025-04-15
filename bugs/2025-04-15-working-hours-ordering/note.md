# Working Hours Ordering
**Date:** 2025-04-15
**Category:** #bug

---

## 🔍 Problem
Availability was missing when working hours were out of order with regards to time.

## 🧠 Context
https://zocdoc.atlassian.net/browse/INTPLAT-2010?linkSource=email

The new working hours page doesn't enforce the ordering.  The old page doesn't allow saving working hours that are out of order.

## 🔬 Investigation
- The sync side algorithm for filtering out timeslots based on working hours depends on the working hours being ordered by ascending time.
- Looked at the sync side code.  Asked Cursor to explain how working hours are used to filter out
time slots.
- Used IPA to examine working hours for providers and practices.
- Sometimes the working hours page sorts the working hours when displaying them, even if they are out of order.  This is because the endpoint to get working hours by practice sorts the working hours when querying them from the DB. I don't quite understand why sometimes it's sorted and other times it's not.
- Used the sync audit tool to test ExtractAvailability on the DLL containing the fix: https://pulse.zocdoc.com/csr/SyncWorkerAudit/DisplayQueryResults?dllId=7582

## ✅ Solution
- Working hours blocks have a `start_date` and `start_time`. We can have two working hours blocks on Monday that have different start dates.  The start date is populated when we first create the working hours.  So let's say we first create a block from 3pm-5pm on Monday, then two weeks later we add another block from 8am-12pm on Monday.  These blocks will both apply to Monday, but their start dates will be different.
- The point above about the start date matters because the sync code was using the entire date time to order working hours blocks when doing its filtering.  It should only use the time to do the ordering correctly.
- https://github.com/Zocdoc/Synchronizers/pull/4485

## 📌 Notes & Takeaways
- Should we look into preventing overlapping blocks? Probably.
