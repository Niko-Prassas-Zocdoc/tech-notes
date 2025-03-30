# Reschedule Events Not Populated
**Date:** 2025-03-30 
**Category:** #bug

---

## 🔍 Problem
A change was rolled out to do atomic reschedules for syncs.  In the post processing of reschedules, we were missing a step to send the appropriate events.  In particular, we were missing sending the `AppointmentRescheduleCreated` event.  AppointmentBox depends on this event to populate the correct in its system.  So AppointmentBox was still showing the original time.

Aside from fixing the issue, we also want to backfill the missing events because a lot of appointments were impacted.

## 🧠 Context
- Due to recent project to due atomic reschedules for syncs, instead of cancel and rebooking.
- Who else was involved? Shashank, Stirpe.
- [Slack channel](https://zocdoc.enterprise.slack.com/archives/C08KKKZ9ZK4)
- Speadsheet of impacted appointments: [sheet](https://docs.google.com/spreadsheets/d/1PzvN-PVHLIfyyaiQfd68YO0tn4xw42YDsQEks_1yiy0/edit?gid=284945391#gid=284945391)

## 🔬 Investigation
[PR](https://github.com/Zocdoc/zocdoc_web/pull/62614) to create the event.  Is this gonna just work?  Using code that Shashank is using to create the events.  It has a lot of logic though. If the data is wonky enough, the events could not get created based on the logic inside the code.

Here is an impacted appointment: `app_000bb655-c2a2-4a22-9c6c-16875727ca92`

### Question 1
For `AppointmentRescheduleCreated` events, is the request id for these events the same as the request id of the corresponding `AppointmentRescheduleRequested` event?  I can probably use Cistern and/or AppointmentLedger to figure this out.

### Answer 1
Seems like yes, it is acceptable for the `AppointmentRescheduleCreated` and `AppointmentRescheduleRequested` to have the same request id, at least according to the following appointment in cistern:
```sql
SELECT APPOINTMENT_ID,NAME,REQUEST_ID,TIMESTAMP_UTC FROM APPOINTMENT.APPOINTMENT_EVENT
where appointment_id = 'app_8b46b248-7310-45b4-a30f-658baaeb7af5'
```

### Question 2
Given a request id, how does the monolith code determine if it's a reschedule or not?  [Relevant code](https://github.com/Zocdoc/zocdoc_web/blob/version_2025-03-28-1500/ZocDoc.Booking/ZocDoc.Booking.Impl/EventRecorder/AppointmentEventRecorder.cs#L394).

### Answer 2
Eventually we get to [this](https://github.com/Zocdoc/zocdoc_web/blob/cc8c18545d9e1f070c4ab4a4bbf5711d06a331d3/ZocDoc.Booking/ZocDoc.Booking.Impl/EventRecorder/BaseEventRecorder.cs#L261) code.  This method populates the `UpdatedExistingAppointment` property that is used [here](https://github.com/Zocdoc/zocdoc_web/blob/version_2025-03-28-1500/ZocDoc.Booking/ZocDoc.Booking.Impl/EventRecorder/AppointmentEventRecorder.cs#L408).  The code seems to indicate that:

UpdatedExistingAppointment will be true if:
    The appointment has already been created (has an AppointmentCreated event) AND either:
    - The appointment is a reschedule (has OldRequestId) OR
    - It's a practice confirmation and the appointment wasn't already confirmed

I am trying to determine if this will be true for the impacted appointments, because I am hoping the existing code will create the events without having to write too much extra code.  I think/hope it will work, because the impacted appointments have been created, and they should have the new request id for the reschedule.  For instance, if I looked at impacted appointment `app_000bb655-c2a2-4a22-9c6c-16875727ca92` in AppointmentLedger, the latest cache appointment has an old request id that is not equal to the current request id.  So hopefully the endpoint I am adding does what it's supposed to do.

I believe for my endpoint, I should pass the request id corresponding to the `AppointmentRescheduleRequested` event.

## ✅ Solution
- What fixed it and why
- Commands or config changes

## 📌 Notes & Takeaways
- The data being correct in the monolith doesn't mean it's gonna be correct in AppointmentBox.  The AppointmentBox and any other external systems need to be tested thoroughly because they may need to populate their own data stores.
