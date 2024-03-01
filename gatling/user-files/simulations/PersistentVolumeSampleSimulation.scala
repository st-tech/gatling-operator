// Sample originally from https://github.com/gatling/gatling/blob/400a125d7995d1b895c4cc4847ff15059d252948/gatling-bundle/src/main/scala/computerdatabase/BasicSimulation.scala
/*
* Copyright 2011-2021 GatlingCorp (https://gatling.io)
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*  http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/
import scala.concurrent.duration._

import io.gatling.core.Predef._
import io.gatling.http.Predef._

class PersistentVolumeSampleSimulation extends Simulation {

  val usersPerSec = sys.env.getOrElse("CONCURRENCY", "1").toInt
  val durationSec = sys.env.getOrElse("DURATION", "10").toInt

  val feeder_user_ids = csv("pv/user_ids.csv")
  val feeder_goods_ids = csv("pv/goods_ids.csv")
  val feeder_myresources = csv("myresources.csv")

  // A scenario is a chain of requests and pauses
  val scn = scenario("Scenario Name")
    .feed(feeder_user_ids.circular)
    .feed(feeder_goods_ids.circular)
    .feed(feeder_myresources.circular)
    .exec { session =>
      println(s"User ${session("user_id").as[String]} is buying goods ${session("goods_id").as[String]} ${session("goods_name").as[String]}")
      println(s"myresource: ${session("alphabet").as[String]}")
      session
    }
    // Note that Gatling has recorded real time pauses
    .pause(1)

  setUp(
    scn.inject(
      constantUsersPerSec(usersPerSec) during(durationSec seconds)
    )
  )
}
