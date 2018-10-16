;;; Copyright (c) 2018
;;; IoTech Ltd
;;; SPDX-License-Identifier: Apache-2.0

(ns org.edgexfoundry.ui.manager.ui.main
  (:require [fulcro.client.primitives :as prim :refer [defui defsc]]
            [fulcro.client.dom :as dom]
            [fulcro.client.routing :as r :refer [defrouter]]
            [org.edgexfoundry.ui.manager.ui.commands :as c]
            [org.edgexfoundry.ui.manager.ui.devices :as d]
            [org.edgexfoundry.ui.manager.ui.readings :as rd]
            [org.edgexfoundry.ui.manager.ui.profiles :as p]
            [org.edgexfoundry.ui.manager.ui.addressables :as a]
            [org.edgexfoundry.ui.manager.ui.common :as co]
            [org.edgexfoundry.ui.manager.ui.ident :as id]
            [org.edgexfoundry.ui.manager.ui.schedules :as sc]
            [org.edgexfoundry.ui.manager.ui.logging :as log]
            [org.edgexfoundry.ui.manager.ui.notifications :as nt]
            [org.edgexfoundry.ui.manager.ui.exports :as ex]))

(defn select-router [props]
  ;(js/console.log "props" props)
  (condp #(contains? %2 %1) props
    :commands co/command-list-ident
    :ui/show-devices co/device-list-ident
    :ui/reading-list co/reading-page-ident
    :ui/show-profiles co/profile-list-ident
    :ui/show-addressables co/addressable-list-ident
    :ui/show-profile-yaml co/profile-yaml-ident
    :ui/show-schedules co/schedules-list-ident
    :ui/show-schedule-events co/schedule-events-list-ident
    :ui/show-logs co/log-entry-list-ident
    :ui/show-notifications co/notifications-list-ident
    :ui/show-exports co/exports-list-ident
    (id/edgex-ident props)))

(defrouter DeviceListOrInfoRouter :device-router
  (fn [this props] (select-router props))
  :show-devices d/DeviceList
  :show-commands c/CommandList
  :show-profiles p/ProfileList
  :show-profile-yaml p/ProfileYAML
  :show-addressables a/AddressableList
  :show-schedules sc/ScheduleList
  :show-schedule-events sc/ScheduleEventList
  :show-logs log/LogEntryList
  :show-notifications nt/NotificationList
  :show-exports ex/ExportList
  :reading-page rd/ReadingsPage
  :device d/DeviceInfo
  :schedule-event sc/ScheduleEventInfo)

(def ui-device-list-or-info (prim/factory DeviceListOrInfoRouter))

(defsc MainPage [this {:keys [device-data]}]
  {:query [{:device-data (prim/get-query DeviceListOrInfoRouter)}]
   :ident (fn [] co/device-page-ident)
   :initial-state (fn [p] {:device-data (prim/get-initial-state DeviceListOrInfoRouter nil)})}
  (dom/div nil
           (ui-device-list-or-info device-data)))

(def ui-main-page (prim/factory MainPage))
